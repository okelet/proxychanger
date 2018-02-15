#!/bin/bash

#COLOREND=$(tput sgr0)
#GREEN=$(tput setaf 2)
#RED=$(tput setaf 1)	
#UNDER=$(tput smul)
#BOLD=$(tput bold)

APP_PATH=${HOME}/tmp/proxychanger

######################################################################

IS_DEBIAN=0
IS_REDHAT=0
if [ -f /etc/debian_version ]; then
    IS_DEBIAN=1
    DEPENDENCIES=(libgtk-3-dev libappindicator3-1 libappindicator3-dev git)
    echo "Debian based system detected."
elif [[ -e /etc/centos-release || -e /etc/redhat-release ]]; then
    IS_REDHAT=1
    DEPENDENCIES=(gtk3-devel libappindicator-gtk3 libappindicator-gtk3-devel git)
    echo "Red Hat based system detected."
else
    echo "Unsupported operating system."
    exit 1
fi

######################################################################

function check_dependency() {
    local DEP=$1
    local IS_DEB=$2
    local IS_RH=$3
    if [ ${IS_DEB} -eq 1 ]; then
        if dpkg-query -W -f'${Status}' ${DEP} 2>/dev/null | grep -q "ok installed"; then
            return 0
        else
            return 1
        fi
    else
        if rpm -q ${DEP} 2>&1 > /dev/null ; then
            return 0
        else
            return 1
    fi
fi
}

function check_and_install_dependency() {
    local PKG=$1
    local IS_DEB=$2
    local IS_RH=$3
    local REPO=$4
    if ! check_dependency ${PKG} ${IS_DEB} ${IS_RH} ; then
        while true; do 
            read -p "${PKG} is not installed; do you want to install it? " -u 3 yn
            case $yn in
                [Yy]* )
                    if [ -n "${REPO}" ]; then
                        if [[ ${IS_DEB} -eq 1 ]]; then
                            sudo add-apt-repository --update --yes ${REPO}
                            RET=$?
                            if [ $RET -ne 0 ]; then
                                return 1
                            fi
                        elif [[ ${IS_RH} -eq 1 ]]; then
                            sudo yum-config-manager --add-repo ${REPO}
                            RET=$?
                            if [ $RET -ne 0 ]; then
                                return 1
                            fi
                        else
                            return 1
                        fi
                    fi
                    if [ ${IS_DEB} -eq 1 ]; then
                        sudo apt-get install ${PKG}
                    elif [ ${IS_RH} -eq 1 ]; then
                        sudo yum install ${PKG}
                    else
                        return 1
                    fi
                    break
                    ;;
                [Nn]* )
                    return 1
                    ;;
                * ) echo "Please answer yes or no.";;
            esac
        done
    fi
    check_dependency ${PKG} ${IS_DEB} ${IS_RH} || return 1
    return 0
}


for i in "${DEPENDENCIES[@]}" ; do
    exec 3<>/dev/tty
    check_and_install_dependency ${i} ${IS_DEBIAN} ${IS_REDHAT}
    RET=$?
    exec 3>&-
    if [ ${RET} -ne 0 ]; then
        echo "Failed to detect or install dependency ${i}."
        exit 1
    fi
done

######################################################################

# Test for some default go installations
if [[ -z "$(which go)" && -e "/usr/lib/go-1.9/bin/go" ]]; then
    export PATH=${PATH}:/usr/lib/go-1.9/bin
    export GOROOT=/usr/lib/go-1.9
fi

if [[ -z "$(which go)" && -e "/usr/lib/go-1.8/bin/go" ]]; then
    export PATH=${PATH}:/usr/lib/go-1.8/bin
    export GOROOT=/usr/lib/go-1.8
fi

if [ -z "$(which go)" ]; then
    exec 3<>/dev/tty
    if [ ${IS_DEBIAN} -eq 1 ]; then
        check_and_install_dependency golang-1.9-go ${IS_DEBIAN} ${IS_REDHAT} "ppa:gophers/archive"
        RET=$?
    else
        check_and_install_dependency golang ${IS_DEBIAN} ${IS_REDHAT}
        RET=$?
    fi
    exec 3>&-
    if [ ${RET} -ne 0 ]; then
        echo "Failed to detect or install golang."
        exit 1
    fi
fi

if [ -z "${GOPATH}" ]; then
    PRJ_GO_PATH=$(dirname $(dirname $(dirname $(dirname $(dirname $(realpath $0))))))
    mkdir -p ${HOME}/tmp/go
    export GOPATH=${HOME}/tmp/go:${PRJ_GO_PATH}
    echo "Detected empty GOPATH; auto-set to ${GOPATH}."
fi

######################################################################

[ ! -e ${HOME}/.proxychanger ] && mkdir ${HOME}/.proxychanger

echo "Generating assets..."
OUT=$(go-bindata -prefix proxychangerlib -pkg proxychangerlib -o proxychangerlib/assets.go proxychangerlib/assets/... 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error generating assets (${RET}): ${OUT}"
    exit 1
fi

echo "Compiling translations..."
for i in $(find locale -name "*.po") ; do
    SRC_DIR=$(dirname "${i}")
    SOURCE=${i}
    DEST="${SRC_DIR}/$(basename "${i}" .po).mo"
    OUT=$(msgfmt "${SOURCE}" -o "${DEST}" 2>&1)
    RET=$?
    if [ ${RET} -ne 0 ]; then
        echo "Error compiling translation for ${SOURCE} (${RET}): ${OUT}"
        exit 1
    fi
done

echo "Installing translations..."
[ -e ${HOME}/.proxychanger/locale ] && rm -Rf ${HOME}/.proxychanger/locale
OUT=$(find locale -name "*.mo" | xargs cp --parents -v -t ~/.proxychanger 2>&1)
RET=$?
if [ $RET -ne 0 ]; then
    echo "Error installing translations (${RET}): ${OUT}"
    exit 1
fi

echo "Installing icon..."
[ ! -f ${HOME}/.local/share/icons ] && mkdir -p ${HOME}/.local/share/icons
OUT=$(cp proxychanger.png ${HOME}/.local/share/icons/proxychanger.png 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error installing desktop shortcut (${RET}): ${OUT}"
    exit 1
fi

echo "Installing applications shortcut..."
SHORTCUT_FILE="${HOME}/.local/share/applications/proxychanger.desktop"
[ ! -e "${HOME}/.local/share/applications" ] && mkdir -p "${HOME}/.local/share/applications"
OUT=$(cp proxychanger.desktop ${SHORTCUT_FILE} 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error installing applications shortcut (${RET}): ${OUT}"
    exit 1
fi
sed -i -e "s|^Exec=.*|Exec=${APP_PATH}|" "${SHORTCUT_FILE}"

echo "Installing auto start shortcut..."
AUTOSTART_FILE="${HOME}/.config/autostart/proxychanger.desktop"
if [ -f "${AUTOSTART_FILE}" ]; then
    AUTOSTART_OPTS="$(egrep -i "^X-GNOME-Autostart-enabled" "${AUTOSTART_FILE}")"
fi
[ ! -e "${HOME}/.config/autostart" ] && mkdir -p "${HOME}/.config/autostart"
OUT=$(cp proxychanger.desktop ${AUTOSTART_FILE} 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error installing auto start shortcut (${RET}): ${OUT}"
    exit 1
fi
if [ -n "${AUTOSTART_OPTS}" ]; then
    echo "${AUTOSTART_OPTS}" >> "${AUTOSTART_FILE}"
fi
sed -i -e "s|^Exec=.*|Exec=${APP_PATH}|" "${AUTOSTART_FILE}"

echo "Compiling application..."
[ ! -r ~/tmp ] && mkdir ~/tmp
OUT=$(go get -tags gtk_$(pkg-config --modversion gtk+-3.0 | tr . _ | cut -d '_' -f 1-2) 2>&1)
RET=$?
if [ $RET -eq 0 ]; then
    OUT=$(go build -o ~/tmp/proxychanger -tags gtk_$(pkg-config --modversion gtk+-3.0 | tr . _| cut -d '_' -f 1-2) 2>&1)
    RET=$?
    if [ $RET -ne 0 ]; then
        echo "Error compiling application (${RET}): ${OUT}"
        exit 1
    fi
else
    echo "Error downloading application dependencies (${RET}): ${OUT}"
    exit 1
fi

echo "Executable generated in ~/tmp/proxychanger."
if [ -n "${RUN}" ]; then
    ~/tmp/proxychanger
fi
