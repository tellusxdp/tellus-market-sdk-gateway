#!/usr/bin/env bash

# error codes
# 0 - exited without problems
# 1 - parameters not supported were used or some unexpected error occurred
# 2 - OS not supported by this script
# 3 - installed version of tellus-market-sdk-gateway is up to date
# 4 - supported unzip tools are not available

set -e

INSTALL_SH_URL="https://raw.githubusercontent.com/tellusxdp/tellus-market-sdk-gateway/master/files/install.sh"

#when adding a tool to the list make sure to also add it's corresponding command further in the script
unzip_tools_list=('unzip' '7z' 'busybox')

usage() { echo "Usage: curl ${INSTALL_SH_URL} | sudo bash" 1>&2; exit 1; }

#create tmp directory and move to it with macOS compatibility fallback
tmp_dir=`mktemp -d 2>/dev/null || mktemp -d -t 'tellus-market-sdk-gateway-install.XXXXXXXXXX'`; cd $tmp_dir


#make sure unzip tool is available and choose one to work with
set +e
for tool in ${unzip_tools_list[*]}; do
    trash=`hash $tool 2>>errors`
    if [ "$?" -eq 0 ]; then
        unzip_tool="$tool"
        break
    fi
done  
set -e

# exit if no unzip tools available
if [ -z "${unzip_tool}" ]; then
    printf "\nNone of the supported tools for extracting zip archives (${unzip_tools_list[*]}) were found. "
    printf "Please install one of them and try again.\n\n"
    exit 4
fi

#detect the platform
OS="`uname`"
case $OS in
  Linux)
    OS='linux'
    ;;
  *)
    echo 'OS not supported'
    exit 2
    ;;
esac

OS_type="`uname -m`"
case $OS_type in
  x86_64|amd64)
    OS_type='amd64'
    ;;
  *)
    echo 'Archtecture not supported'
    exit 2
    ;;
esac

#download and unzip
zipfile="tellus-market-sdk-gateway-$OS-$OS_type.zip"
download_link="https://github.com/tellusxdp/tellus-market-sdk-gateway/releases/latest/download/$zipfile"

curl -L -O $download_link
unzip_dir="tmp_unzip_dir_for_tellus-market-sdk-gateway"
# there should be an entry in this switch for each element of unzip_tools_list
case $unzip_tool in
  'unzip')
    unzip -a $zipfile -d $unzip_dir
    ;;
  '7z')
    7z x $zipfile -o$unzip_dir
    ;;
  'busybox')
    mkdir -p $unzip_dir
    busybox unzip $zipfile -d $unzip_dir
    ;;
esac
    
cd $unzip_dir


#mounting tellus-market-sdk-gateway to environment

case $OS in
  'linux')
    #binary
    cp tellus-market-sdk-gateway /usr/bin/tellus-market-sdk-gateway.new
    chmod 755 /usr/bin/tellus-market-sdk-gateway.new
    chown root:root /usr/bin/tellus-market-sdk-gateway.new
    mv /usr/bin/tellus-market-sdk-gateway.new /usr/bin/tellus-market-sdk-gateway

    #config
    mkdir -p /etc/gateway
    if [ -e /etc/gateway/config.yml ]; then
      printf "\config file is already exist."
    else
      cp config.yml /etc/gateway/config.yml
      chmod 644 /etc/gateway/config.yml
      chown root:root /etc/gateway/config.yml
    fi

    #systemd service
    cp gateway.service /tmp/gateway.service
    chmod 644 /tmp/gateway.service
    chown root:root /tmp/gateway.service
    mv /tmp/gateway.service /etc/systemd/system/gateway.service
    systemctl daemon-reload
    systemctl enable gateway
    ;;
  *)
    echo 'OS not supported'
    exit 2
esac


#update version variable post install
version=`tellus-market-sdk-gateway --version 2>>errors | head -n 1`

#create the directory for the default configuration folder
mkdir -p /var/lib/gateway/autocert

printf "\nsuccessfully installed."
exit 0
