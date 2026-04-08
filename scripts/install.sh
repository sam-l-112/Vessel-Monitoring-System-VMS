#!/usr/bin/env bash
set -euo pipefail

# install.sh
# 安裝 Ansible 自動化部署前所需工具

# 目前腳本僅支援 Debian / Ubuntu 系統

function command_exists() {
  command -v "$1" >/dev/null 2>&1
}

function install_packages() {
  local pkgs=(python3 python3-pip ssh git ansible docker.io docker-compose nginx mariadb-server nodejs npm)
  echo "[INFO] Installing packages: ${pkgs[*]}"
  sudo apt-get update
  sudo apt-get install -y "${pkgs[@]}"
}

function install_ansible_via_pip() {
  if ! command_exists ansible; then
    echo "[INFO] Installing Ansible via pip3"
    sudo pip3 install ansible
  fi
}

function install_docker_compose_via_pip() {
  if ! command_exists docker-compose; then
    echo "[INFO] Installing docker-compose via pip3"
    sudo pip3 install docker-compose
  fi
}

function main() {
  if [[ "$EUID" -ne 0 ]]; then
    echo "[WARN] 建議使用 sudo 執行此腳本。"
  fi

  if ! command_exists apt-get; then
    echo "[ERROR] 目前僅支援 Debian/Ubuntu 系統。"
    exit 1
  fi

  install_packages
  install_ansible_via_pip
  install_docker_compose_via_pip

  echo "[INFO] 安裝完成。"
  echo "請確認以下服務已啟動或可用："
  echo "  - ansible"
  echo "  - docker"
  echo "  - docker-compose"
  echo "  - nginx"
  echo "  - mariadb"
  echo "  - nodejs"
  echo "  - python3"
}

main "$@"
