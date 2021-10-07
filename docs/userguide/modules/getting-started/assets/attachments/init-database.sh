#!/bin/bash

export pgo_cluster_name=hippo
export cluster_namespace=my-postgresql
export pgo_cluster_username=hippo
export PGPASSWORD=$(kubectl -n "${cluster_namespace}" get secrets \
  "${pgo_cluster_name}-pguser-${pgo_cluster_username}" -o "jsonpath={.data['password']}" | base64 -d)
nohup kubectl -n ${cluster_namespace} port-forward svc/hippo-pgbouncer 5432:5432 &
sleep 5
curl -LO https://raw.githubusercontent.com/spring-petclinic/spring-petclinic-rest/master/src/main/resources/db/postgresql/initDB.sql
psql -h localhost -U "${pgo_cluster_username}" "${pgo_cluster_name}" -f initDB.sql
curl -LO https://raw.githubusercontent.com/spring-petclinic/spring-petclinic-rest/master/src/main/resources/db/postgresql/populateDB.sql
psql -h localhost -U "${pgo_cluster_username}" "${pgo_cluster_name}" -f populateDB.sql
