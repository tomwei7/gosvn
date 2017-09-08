#!/bin/sh
set -e
SVNROOT=/svnroot
SVNREPONAME=test_svn_repo
SVNREPO=${SVNROOT}/${SVNREPONAME}
export SVNUSER=testuser
export SVNPASSWD=testpasswd
mkdir -p ${SVNROOT}
svnserve --log-file ${SVNROOT}/svnserve.log -d -r ${SVNROOT}
svnadmin create ${SVNREPO}
echo "${SVNUSER} = ${SVNPASSWD}" >> ${SVNREPO}/conf/passwd
sed -i "s/# password-db = passwd/password-db = passwd/g" ${SVNREPO}/conf/svnserve.conf
export TEST_SVNURL="svn://${SVNUSER}:${SVNPASSWD}@127.0.0.1:3690/${SVNREPONAME}"
exec "$@"
