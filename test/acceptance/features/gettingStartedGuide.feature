@getting-started
Feature: Getting Started Guide

  Make sure the scenarios included in the Getting Started Guide works.

  Background:
    Given Namespace [TEST_NAMESPACE] is used
    * Service Binding Operator is running

  @external-feedback
  Scenario: Connecting PetClinic application to PostgreSQL database
    Given PetClinic sample application is installed
    * PostgreSQL database is running
    When Service Binding is applied from "samples/apps/spring-petclinic/petclinic-postgresql-binding.yaml" file
    Then Service Binding is ready
    * Service Binding secret contains "username" key
    * Service Binding secret contains "password" key
    * Service Binding secret contains "database" key
    * Service Binding secret contains "host" key
    * Service Binding secret contains "port" key
    * Service Binding secret contains "type" key
    * Service Binding is deleted

  @olm
  @disable.arch.ppc64le
  @disable.arch.s390x
  @disable.arch.arm64
  Scenario: Connecting PetClinic application to an Operator-backed PostgreSQL database
    Given Crunchy Data Postgres operator is running
    * PetClinic sample application is installed
    * PostgresCluster database is running
    When Service Binding is applied from "samples/apps/spring-petclinic/petclinic-pgcluster-binding.yaml" file
    Then Service Binding is ready
    * Service Binding secret contains "username" key
    * Service Binding secret contains "password" key
    * Service Binding secret contains "database" key
    * Service Binding secret contains "host" key
    * Service Binding secret contains "port" key
    * Service Binding secret contains "type" key
    * Service Binding is deleted