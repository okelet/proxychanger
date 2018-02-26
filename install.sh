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
elif [[ -e /etc/centos-release || -e /etc/redhat-release ]]; then
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

if [[ "${IGNORE_DEPS}" != "1" ]]; then
    for i in "${DEPENDENCIES[@]}" ; do
        if ! check_dependency ${i} ${IS_DEBIAN} ${IS_REDHAT} ; then
            echo "Missing dependency ${i}."
            echo "Please install it using the command below and re-run this script."
            if [[ ${IS_DEBIAN} -eq 1 ]]; then
                echo "sudo apt-get install ${i}"
            elif [[ ${IS_REDHAT} -eq 1 ]] ; then
                echo "sudo yum install ${i}"
            fi
            exit 1
        fi
    done
fi


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
