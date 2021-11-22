#!/bin/bash

export PGHOST=
export PGPORT=
export PGDATABASE=
export PGUSER=
export PGPASSWORD=

# Проверяем версию
V=$(psql -t -A -c "select (regexp_match(obj_description('reporter'::regnamespace),'\d+\.\d+\.\d+'))[1];")
echo "Текущая версия: $V"
for F in db/*.sql; do
  echo -n "${F} ..."
  FV=$(basename ${F%%\.sql})
  FV=${FV//_/}
  if [[ "$FV" > "${V//\./}" ]]; then
    echo "исполняется"
    psql -1 -f "$F"
  else
    echo "пропущен"
  fi
done