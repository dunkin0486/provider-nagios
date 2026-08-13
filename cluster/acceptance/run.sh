#!/usr/bin/env bash
set -euo pipefail

projectdir="$(cd "$(dirname "${BASH_SOURCE[0]}")"/../.. && pwd)"
compose_file="${projectdir}/cluster/acceptance/docker-compose.yaml"
env_file="${projectdir}/cluster/acceptance/.env"

if [[ -f "${env_file}" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "${env_file}"
  set +a
fi

is_true() {
  case "${1:-}" in
    1|true|TRUE|yes|YES|on|ON)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

get_nagios_api_token() {
  local service="${1:-nagiosxi}"
  local cfg_path="/opt/nagiosxi/nagiosxi/basedir/html/config.inc.php"
  local token
  local db_user
  local db_pass
  local db_name
  local db_host

  token="$(docker compose -f "${compose_file}" exec -T "${service}" bash -lc '
    set -e
    if [[ -f /opt/nagiosxi/xi-sys.cfg ]]; then
      . /opt/nagiosxi/xi-sys.cfg
      mysql -u nagiosxi -p"$nagiosxipass" nagiosxi -e "UPDATE xi_users SET api_enabled=1 WHERE username=\"nagiosadmin\";" >/dev/null 2>&1 || true
      mysql -u nagiosxi -p"$nagiosxipass" nagiosxi -N -e "SELECT api_key FROM xi_users WHERE username=\"nagiosadmin\" LIMIT 1;" 2>/dev/null | tr -d "\r" | tail -n1
    fi
  ' 2>/dev/null || true)"
  if [[ "${token}" =~ ^[A-Za-z0-9]{16,}$ ]]; then
    printf '%s\n' "${token}"
    return 0
  fi

  db_user="$(docker compose -f "${compose_file}" exec -T "${service}" php -r '$cfg=[]; include "'"${cfg_path}"'"; echo $cfg["db_info"]["nagiosxi"]["user"] ?? "";')"
  db_pass="$(docker compose -f "${compose_file}" exec -T "${service}" php -r '$cfg=[]; include "'"${cfg_path}"'"; echo $cfg["db_info"]["nagiosxi"]["pwd"] ?? "";')"
  db_name="$(docker compose -f "${compose_file}" exec -T "${service}" php -r '$cfg=[]; include "'"${cfg_path}"'"; echo $cfg["db_info"]["nagiosxi"]["db"] ?? "nagiosxi";')"
  db_host="$(docker compose -f "${compose_file}" exec -T "${service}" php -r '$cfg=[]; include "'"${cfg_path}"'"; echo $cfg["db_info"]["nagiosxi"]["dbserver"] ?? "";')"

  if [[ -z "${db_host}" ]]; then
    db_host="localhost"
  fi

  docker compose -f "${compose_file}" exec -T "${service}" mysql -h "${db_host}" -u "${db_user}" -p"${db_pass}" "${db_name}" -e \
    "UPDATE xi_users SET api_enabled=1 WHERE username='nagiosadmin';" >/dev/null 2>&1

  docker compose -f "${compose_file}" exec -T "${service}" mysql -h "${db_host}" -u "${db_user}" -p"${db_pass}" "${db_name}" -N -e \
    "SELECT api_key FROM xi_users WHERE username='nagiosadmin' LIMIT 1;" 2>/dev/null | tr -d '\r' | tail -n1
}

wait_for_nagios() {
  local service="${1:-nagiosxi}"
  local timeout_secs="${2:-3600}"
  local elapsed=0
  local step=5
  local cid
  local running

  while (( elapsed < timeout_secs )); do
    cid="$(docker compose -f "${compose_file}" ps -q "${service}" | head -n1)"
    if [[ -z "${cid}" ]]; then
      sleep "${step}"
      elapsed=$((elapsed + step))
      continue
    fi

    running="$(docker inspect -f '{{.State.Running}}' "${cid}" 2>/dev/null || echo false)"

    if [[ "${running}" == "true" ]]; then
      return 0
    fi

    sleep "${step}"
    elapsed=$((elapsed + step))
  done

  echo "timeout waiting for ${service} to be ready"
  docker compose -f "${compose_file}" logs --tail=200 "${service}" || true
  return 1
}

wait_for_nagios_install() {
  local service="${1:-nagiosxi}"
  local timeout_secs="${2:-7200}"
  local elapsed=0
  local step=10
  local cid
  local running
  local state
  local installed
  local marker_status
  local latest_install_log

  while (( elapsed < timeout_secs )); do
    cid="$(docker compose -f "${compose_file}" ps -q "${service}" | head -n1)"
    running="false"
    if [[ -n "${cid}" ]]; then
      running="$(docker inspect -f '{{.State.Running}}' "${cid}" 2>/dev/null || echo false)"
    fi

    if [[ "${running}" != "true" ]]; then
      printf '[acceptance] nagios install wait: elapsed=%ss state=%s markers="%s"\n' "${elapsed}" "not-running" "unavailable"
      sleep "${step}"
      elapsed=$((elapsed + step))
      continue
    fi

    state="$(docker compose -f "${compose_file}" exec -T "${service}" bash -lc 'systemctl is-active nagiosxi-install.service 2>/dev/null || true' 2>/dev/null | tr -d '\r' || true)"
    installed="$(docker compose -f "${compose_file}" exec -T "${service}" bash -lc 'test -f /opt/nagiosxi/installed.nagiosxi && echo yes || echo no' 2>/dev/null | tr -d '\r' || true)"

    marker_status="$(docker compose -f "${compose_file}" exec -T "${service}" bash -lc 'for f in /opt/nagiosxi/installed.mysql /opt/nagiosxi/installed.nagios /opt/nagiosxi/installed.nagiosxi /opt/nagiosxi/installed.plugins; do if [ -f "$f" ]; then printf "%s=ok " "$(basename "$f")"; else printf "%s=pending " "$(basename "$f")"; fi; done' 2>/dev/null | tr -d '\r' || true)"
    latest_install_log="$(docker compose -f "${compose_file}" exec -T "${service}" bash -lc 'journalctl -u nagiosxi-install.service --no-pager -n 1 2>/dev/null | sed -E "s/[[:space:]]+/ /g"' 2>/dev/null | tr -d '\r' || true)"

    if [[ -z "${state}" ]]; then
      state="unavailable"
    fi
    if [[ -z "${installed}" ]]; then
      installed="no"
    fi
    if [[ -z "${marker_status}" ]]; then
      marker_status="unavailable"
    fi

    printf '[acceptance] nagios install wait: elapsed=%ss state=%s markers="%s"\n' "${elapsed}" "${state:-unknown}" "${marker_status}"
    if [[ -n "${latest_install_log}" ]]; then
      printf '[acceptance] nagios install latest: %s\n' "${latest_install_log}"
    fi

    if [[ "${installed}" == "yes" && "${state}" != "activating" ]]; then
      return 0
    fi

    sleep "${step}"
    elapsed=$((elapsed + step))
  done

  echo "timeout waiting for Nagios XI install to complete"
  docker compose -f "${compose_file}" exec -T "${service}" bash -lc 'systemctl status nagiosxi-install.service --no-pager || true' || true
  docker compose -f "${compose_file}" exec -T "${service}" bash -lc 'journalctl -u nagiosxi-install.service --no-pager -n 200 || true' || true
  return 1
}

wait_for_api_token() {
  local service="${1:-nagiosxi}"
  local timeout_secs="${2:-3600}"
  local elapsed=0
  local step=10
  local token

  while (( elapsed < timeout_secs )); do
    token="$(get_nagios_api_token "${service}" 2>/dev/null || true)"
    if [[ "${token}" =~ ^[A-Za-z0-9]{16,}$ ]]; then
      printf '%s\n' "${token}"
      return 0
    fi

    printf '[acceptance] waiting for api token: elapsed=%ss\n' "${elapsed}"

    sleep "${step}"
    elapsed=$((elapsed + step))
  done

  return 1
}

if [[ -z "${NAGIOS_XI_IMAGE:-}" ]]; then
  echo "NAGIOS_XI_IMAGE must be set (for example: ghcr.io/your-org/nagiosxi:latest)"
  exit 1
fi

cleanup() {
  if is_true "${ACCEPTANCE_PRESERVE_STACK:-false}"; then
    echo "[acceptance] preserving compose stack for reuse (ACCEPTANCE_PRESERVE_STACK=true)"
    return 0
  fi

  if is_true "${ACCEPTANCE_KEEP_VOLUMES:-false}"; then
    echo "[acceptance] tearing down containers and network, keeping volumes"
    docker compose -f "${compose_file}" down --remove-orphans
    return 0
  fi

  docker compose -f "${compose_file}" down --volumes --remove-orphans
}

trap cleanup EXIT

if is_true "${ACCEPTANCE_PRESERVE_STACK:-false}"; then
  echo "[acceptance] fast mode enabled: stack will remain running after tests"
elif is_true "${ACCEPTANCE_KEEP_VOLUMES:-false}"; then
  echo "[acceptance] fast mode enabled: volumes will be preserved on teardown"
fi

docker compose -f "${compose_file}" up -d k3s nagiosxi provider

echo "Waiting for nagiosxi container to be running..."
wait_for_nagios nagiosxi 3600

echo "Waiting for Nagios XI install to complete (first run may take a long time)..."
wait_for_nagios_install nagiosxi 7200

if [[ -z "${NAGIOS_API_TOKEN:-}" ]]; then
  echo "NAGIOS_API_TOKEN not set, deriving token from running nagiosxi container"
  NAGIOS_API_TOKEN="$(wait_for_api_token nagiosxi 3600 || true)"
fi

if [[ -z "${NAGIOS_API_TOKEN}" ]]; then
  echo "failed to determine NAGIOS_API_TOKEN"
  exit 1
fi

docker compose -f "${compose_file}" run --rm -e "NAGIOS_API_TOKEN=${NAGIOS_API_TOKEN}" tests
