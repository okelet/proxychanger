#!/bin/bash

#COLOREND=$(tput sgr0)
#GREEN=$(tput setaf 2)
#RED=$(tput setaf 1)	
#UNDER=$(tput smul)
#BOLD=$(tput bold)

REPO_BASE=https://github.com/okelet/proxychanger
APP_PATH=${HOME}/.local/bin/proxychanger

######################################################################

IS_DEBIAN=0
IS_REDHAT=0
if [ -f /etc/debian_version ]; then
    IS_DEBIAN=1
    DEPENDENCIES=(jq libappindicator3-1 curl)
    echo "Debian based system detected."
elif [ -f /etc/redhat-release ]; then
    IS_REDHAT=1
    DEPENDENCIES=(jq libappindicator-gtk3 curl)
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

# Delete old translations
[ -e ${HOME}/.proxychanger/locale ] && rm -Rf ${HOME}/.proxychanger/locale

# Shortcuts paths
SHORTCUT_FILE="${HOME}/.local/share/applications/proxychanger.desktop"
AUTOSTART_FILE="${HOME}/.config/autostart/proxychanger.desktop"

# Get options of current autostart
if [ -f "${AUTOSTART_FILE}" ]; then
    AUTOSTART_OPTS="$(egrep -i "^X-GNOME-Autostart-enabled" "${AUTOSTART_FILE}")"
fi

echo "Installing application..."

# Get releases information
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

# Extract the tag of the version
VERSION=$(jq -e -r '.tag_name' <(echo "${RELEASES_DATA}") 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error extracting version information (${RET}): ${VERSION}"
    exit 1
fi

# Extract the download URL for the file proxychanger_inst.tar.gz, that is generated and uploaded to Github by Travis
URL=$(jq -e -r '.assets[] | select(.name == "proxychanger_inst.tar.gz") | .browser_download_url' <(echo "${RELEASES_DATA}") 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    echo "Error extracting download information (${RET}): ${URL}"
    exit 1
fi

# Generate a random unique name
TMP_FILE=$(mktemp --suffix .tar.gz)

# Download the release file
OUT=$(curl -sSfL -o "${TMP_FILE}" "${URL}" 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    rm -f "${TMP_FILE}"
    echo "Error downloading installer (${RET}): ${OUT}"
    exit 1
fi

# Extract the downloaded URL
OUT=$(tar zxf "${TMP_FILE}" -C "${HOME}" 2>&1)
RET=$?
if [ ${RET} -ne 0 ]; then
    rm -f "${TMP_FILE}"
    echo "Error extracting installer (${RET}): ${OUT}"
    exit 1
fi

# Delete temporary file
rm -f "${TMP_FILE}"

# Fix shortcuts paths
sed -i -e "s|^Exec=.*|Exec=${APP_PATH}|" "${SHORTCUT_FILE}"
sed -i -e "s|^Exec=.*|Exec=${APP_PATH}|" "${AUTOSTART_FILE}"

# Add options to autostart
if [ -n "${AUTOSTART_OPTS}" ]; then
    echo "${AUTOSTART_OPTS}" >> "${AUTOSTART_FILE}"
fi

# Ensure executable permissions
chmod +x ${APP_PATH}

echo "Installation OK (version ${VERSION})."
