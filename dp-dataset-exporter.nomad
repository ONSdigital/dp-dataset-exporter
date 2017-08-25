job "dp-dataset-exporter" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "service"

  group "publishing" {
    count = "{{PUBLISHING_TASK_COUNT}}"

    constraint {
      attribute = "${node.class}"
      value     = "publishing"
    }

    task "dp-dataset-exporter" {
      driver = "exec"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{BUILD_BUCKET}}/dp-dataset-exporter/{{REVISION}}.tar.gz"
      }

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/{{DEPLOYMENT_BUCKET}}/dp-dataset-exporter/{{REVISION}}.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

        args    = [
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