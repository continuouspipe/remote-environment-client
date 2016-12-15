#!/bin/bash

set -e

if [ -L "$0" ] ; then
    DIR="$(dirname "$(readlink -f "$0")")" ;
else
    DIR="$(dirname "$0")" ;
fi

docker pull koalaman/shellcheck

find "$DIR" -type f ! -path "*.git/*" -perm +111 -or -name "*.sh" | while read -r script; do
  echo "Linting '$script':";
  docker run --rm -i koalaman/shellcheck --exclude SC1091 - < "$script";
  echo
done
