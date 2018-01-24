#!/bin/bash

LOCALE_DIR=$(dirname $0)/locale
[ ! -d "${LOCALE_DIR}" ] && mkdir "${LOCALE_DIR}"

DOMAIN=proxychanger

POT_FILE=${LOCALE_DIR}/${DOMAIN}.pot
CODE_POT_FILE=${LOCALE_DIR}/proxychangercode.pot
GLADE_POT_FILE=${LOCALE_DIR}/proxychangerglade.pot

go-xgettext -k MyGettextv -o ${CODE_POT_FILE} main.go proxychangerlib/*.go

xgettext -d ${DOMAIN} -L glade -o ${GLADE_POT_FILE} proxychangerlib/assets/*.glade

msgcat -o ${POT_FILE} --use-first -t utf-8 ${GLADE_POT_FILE} ${CODE_POT_FILE}

for i in $(find ${LOCALE_DIR} -mindepth 1 -maxdepth 1 -type d); do

	DIR_LANG=$(basename ${i})
    LC_MESSAGES_DIR=${LOCALE_DIR}/${DIR_LANG}/LC_MESSAGES
    if [ ! -e ${LC_MESSAGES_DIR} ]; then
        mkdir ${LC_MESSAGES_DIR}
    fi
    PO_FILE=${LC_MESSAGES_DIR}/${DOMAIN}.po
    if [ ! -f  ${PO_FILE} ] ; then
        msginit --no-translator -l ${DIR_LANG} -i ${POT_FILE} -o ${PO_FILE}
    else
        msgmerge --update ${PO_FILE} ${POT_FILE}
    fi

done
