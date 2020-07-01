#!/bin/bash
#-------------------------------------------------------------
# 2020.07.01 (wu) Github-Version setzen
#-------------------------------------------------------------

ver=$1

over=$(git tag -l | tail -1)
ver=${ver[@]/v/}

echo "oldVersion: $over"
echo "newVersion: v$ver"


over=${over[@]/v/}
over=${over[@]/./}
over=${over[@]/./}

nver=${ver[@]/./}
nver=${nver[@]/./}

if [ $nver -gt $over ]; then
  echo "version update"

cdir=$(pwd)
mdir=/home/master/dev/go/src
gdir=https://proxy.golang.org

echo $cdir
echo ${cdir[@]/$mdir/$gdir}

url=${cdir[@]/$mdir/$gdir}/@v/v$ver.info

git tag v$ver
git push origin v$ver
curl $url
else
echo "no version update"
fi

echo


