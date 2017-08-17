job "dp-dataset-exporter" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "service"

  group "web" {
    count = {{WEB_TASK_COUNT}}

    constraint {
      attribute = "${node.class}"
      value     = "web"
    }

    task "dp-dataset-exporter" {
      driver = "exec"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/ons-dp-deployments/dp-dataset-exporter/latest.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

         args = [
                  "${NOMAD_TASK_DIR}/dp-dataset-exporter",
                ]
      }

      service {
        name = "dp-dataset-exporter"
        port = "http"
        tags = ["web"]
      }

      resources {
        cpu    = "{{WEB_RESOURCE_CPU}}"
        memory = "{{WEB_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-dataset-exporter"]
      }
    }
  }

  group "publishing" {
    count = {{PUBLISHING_TASK_COUNT}}

    constraint {
      attribute = "${node.class}"
      value     = "publishing"
    }

    task "dp-dataset-exporter" {
      driver = "exec"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/ons-dp-deployments/dp-dataset-exporter/latest.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

         args = [
                  "${NOMAD_TASK_DIR}/dp-dataset-exporter",
                ]
      }

      service {
        name = "dp-dataset-exporter"
        port = "http"
        tags = ["publishing"]
      }

      resources {
        cpu    = "{{PUBLISHING_RESOURCE_CPU}}"
        memory = "{{PUBLISHING_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-dataset-exporter"]
      }
    }
  }

}