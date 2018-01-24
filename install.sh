#!/bin/bash

#COLOREND=$(tput sgr0)
#GREEN=$(tput setaf 2)
#RED=$(tput setaf 1)	
#UNDER=$(tput smul)
#BOLD=$(tput bold)

REPO_BASE=https://github.com/okelet/proxychanger
APP_PATH=${HOME}/.local/bin/proxychanger

IS_DEBIAN=0
IS_REDHAT=0
if [ -f /etc/debian_version ]; then
    IS_DEBIAN=1
    DEPENDENCIES=(subversion jq libappindicator3-1 python)
    echo "Debian based system detected."
elif [ -f /etc/redhat-release ]; then
    IS_REDHAT=1
    DEPENDENCIES=(subversion jq libappindicator-gtk3 python)
    echo "Red Hat based system detected."
else
    echo "Unsupported operating system."
    exit 1
fi

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
                            sudo http_proxy=$http_proxy https_proxy=$https_proxy no_proxy=$no_proxy add-apt-repository --update --yes ${REPO}
                            RET=$?
                            if [ $RET -ne 0 ]; then
                                return 1
                            fi
                        elif [[ ${IS_RH} -eq 1 ]]; then
                            sudo http_proxy=$http_proxy https_proxy=$https_proxy no_proxy=$no_proxy yum-config-manager --add-repo ${REPO}
                            RET=$?
                            if [ $RET -ne 0 ]; then
                                return 1
                            fi
                        else
                            return 1
                        fi
                    fi
                    if [ ${IS_DEB} -eq 1 ]; then
                        sudo http_proxy=$http_proxy https_proxy=$https_proxy no_proxy=$no_proxy apt-get install ${PKG}
                    elif [ ${IS_RH} -eq 1 ]; then
                        sudo http_proxy=$http_proxy https_proxy=$https_proxy no_proxy=$no_proxy yum install ${PKG}
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
# Parse environment proxy url for subversion
######################################################################

s_proxy=""
if [ -n "${http_proxy}" ]; then s_proxy=${http_proxy} ; fi
if [[ -z ${s_proxy} && -n "${https_proxy}" ]]; then s_proxy=${https_proxy} ; fi

svn_params=""
if [ -n "${s_proxy}" ]; then
    
    p_host=$(python2 -c 'import sys ; from urlparse import urlparse ; print(urlparse(sys.argv[1])).hostname' "${s_proxy}" 2>&1)
    RET=$?
    if [ ${RET} -ne 0 ]; then
        echo "Error parsing proxy URL ${s_proxy} (${RET}): ${p_host}"
        exit 1
    fi
    if [ -n "${p_host}" ]; then
    	svn_params+=" --config-option servers:global:http-proxy-host=${p_host}"
	fi
    
    p_port=$(python2 -c 'import sys ; from urlparse import urlparse ; print(urlparse(sys.argv[1])).port or 8080' "${s_proxy}" 2>&1)
    RET=$?
    if [ ${RET} -ne 0 ]; then
        echo "Error parsing proxy URL ${s_proxy} (${RET}): ${p_port}"
        exit 1
    fi
    if [ -n "${p_port}" ]; then
    	svn_params+=" --config-option servers:global:http-proxy-port=${p_port}"
	fi
    
    p_user=$(python2 -c 'import sys ; from urlparse import urlparse ; print(urlparse(sys.argv[1])).username or ""' "${s_proxy}" 2>&1)
    RET=$?
    if [ ${RET} -ne 0 ]; then
        echo "Error parsing proxy URL ${s_proxy} (${RET}): ${p_user}"
        exit 1
    fi
    if [ -n "${p_user}" ]; then
    	svn_params+=" --config-option servers:global:http-proxy-username=${p_user}"
	fi
    
    p_pass=$(python2 -c 'import sys ; from urlparse import urlparse ; print(urlparse(sys.argv[1])).password or ""' "${s_proxy}" 2>&1)
    RET=$?
    if [ ${RET} -ne 0 ]; then
        echo "Error parsing proxy URL ${s_proxy} (${RET}): ${p_pass}"
        exit 1
    fi
    if [ -n "${p_pass}" ]; then
    	svn_params+=" --config-option servers:global:http-proxy-password=${p_pass}"
	fi
    
fi

######################################################################

[ ! -e ${HOME}/.proxychanger ] && mkdir ${HOME}/.proxychanger

echo "Installing translations..."
[ -e ${HOME}/.proxychanger/locale ] && rm -Rf ${HOME}/.proxychanger/locale
OUT=$(svn export ${svn_params} ${REPO_BASE}/trunk/locale ${HOME}/.proxychanger/locale 2>&1)
RET=$?
if [ $RET -ne 0 ]; then
    echo "Error installing translations (${RET}): ${OUT}"
    exit 1
fi

echo "Installing icon..."
[ ! -f ${HOME}/.local/share/icons ] && mkdir -p ${HOME}/.local/share/icons
OUT=$(curl -sSfL -o ${HOME}/.local/share/icons/proxychanger.png ${REPO_BASE}/raw/master/proxychanger.png 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error installing desktop shortcut (${RET}): ${OUT}"
    exit 1
fi

echo "Installing applications shortcut..."
SHORTCUT_FILE="${HOME}/.local/share/applications/proxychanger.desktop"
[ ! -e "$(dirname "${SHORTCUT_FILE}")" ] && mkdir -p "$(dirnam "${SHORTCUT_FILE}")"
OUT=$(curl -sSfL -o "${SHORTCUT_FILE}" ${REPO_BASE}/raw/master/proxychanger.desktop 2>&1)
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
[ ! -e "$(dirname "${AUTOSTART_FILE}")" ] && mkdir -p "$(dirname "${AUTOSTART_FILE}")"
OUT=$(curl -sSfL -o ${HOME}/.config/autostart/proxychanger.desktop ${REPO_BASE}/raw/master/proxychanger.desktop 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error installing auto start shortcut (${RET}): ${OUT}"
    exit 1
fi
if [ -n "${AUTOSTART_OPTS}" ]; then
    echo "${AUTOSTART_OPTS}" >> "${AUTOSTART_FILE}"
fi
sed -i -e "s|^Exec=.*|Exec=${APP_PATH}|" "${AUTOSTART_FILE}"

echo "Installing application..."
RELEASES_DATA=$(curl -sSfL https://api.github.com/repos/okelet/proxychanger/releases/latest)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error downloading version information (${RET}): ${RELEASES_DATA}"
    exit 1
fi

# Display the contents if running with -x
if [[ "$-" = *"x"* ]]; then
    echo "${RELEASES_DATA}"
fi

VERSION=$(jq -e -r '.tag_name' <(echo "${RELEASES_DATA}") 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error extracting version information (${RET}): ${VERSION}"
    exit 1
fi

URL=$(jq -e -r '.assets[] | select(.name == "proxychanger") | .browser_download_url' <(echo "${RELEASES_DATA}") 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error extracting download information (${RET}): ${URL}"
    exit 1
fi

[ ! -e "$(dirname "${APP_PATH}")" ] && mkdir -p "$(dirname "${APP_PATH}")"
OUT=$(curl -sSfL -o ${APP_PATH} ${URL} 2>&1)
RET=$?
if [ ${RET} -eq 23 ]; then
    echo "Error download application (${RET}): ${OUT}"
    echo "Is the application running? You should stop it."
    exit 1
elif [ ${RET} -ne 0 ]; then
    echo "Error download application (${RET}): ${OUT}"
    exit 1
fi

chmod +x ${APP_PATH}
echo "Installation OK (version ${VERSION})."
