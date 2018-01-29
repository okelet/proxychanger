#!/bin/bash

set -e

if [ -z "${GTK_VERSION}" ]; then
    export GTK_VERSION=$(pkg-config --modversion gtk+-3.0 | tr . _ | cut -d '_' -f 1-2)
fi

export BUILD_TAG=${TRAVIS_TAG} 
if [ -z "${BUILD_TAG}" ]; then
    # When building locally
    if [ -z "${TRAVIS_COMMIT}" ]; then
        export TRAVIS_COMMIT=$(git log --format="%h" -n 1 2>&1)
        RET=$?
        if [ ${RET} -ne 0 ]; then
            echo "Error getting TRAVIS_COMMIT (${RET}): ${TRAVIS_COMMIT}"
            exit 1
        fi
    fi
    export BUILD_TAG="SNAP-$(date +%Y%m%d%H%M%S)-${TRAVIS_COMMIT}"
fi

echo "Using GTK version ${GTK_VERSION}"
echo "Version is ${BUILD_TAG}"

go build -tags gtk_${GTK_VERSION} -ldflags "-X main.Version=${BUILD_TAG}"

mkdir -p .local/share/icons .local/share/applications .config/autostart .local/bin .proxychanger
cp proxychanger.png .local/share/icons/
cp proxychanger.desktop .local/share/applications/
cp proxychanger.desktop .config/autostart/
cp proxychanger .local/bin

for i in $(find locale -name "*.po") ; do
    SRC_DIR=$(dirname "${i}")
    SOURCE=${i}
    DEST="${SRC_DIR}/$(basename "${i}" .po).mo"
    msgfmt "${SOURCE}" -o "${DEST}"
done

find locale -name "*.mo" | xargs cp --parents -t .proxychanger

tar czvf proxychanger_inst.tar.gz .local .config .proxychanger
