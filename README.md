# Prometheus-Mailgun-Events-Exporter

Mailgun Events Prometheus exporter

## Description

This application is a prometheus exporter for mailgun Events, which aims to monitor the following things:

* `email_delivery_time_seconds`
* `email_delivery_error_messages`

## Authentication

Authentication towards the Mailgun API is being done with exp two ways:
To authenticate with Mailgun API, you need to set `MG_API_KEY`

## List of available metrics

```md
# HELP mailgun_delivery_time_seconds The time took for an email to actually got delivered from the time that got accepted in mailgun
# HELP maigun_delivery_error Email Delivery error messages
```

## Release

The repository has automated builds configured in the DockerHub, for `main` branch and `latest` docker tag.

## How to pull the exporter

```sh
    docker pull gathertown/prometheus-mailgun-events-exporter:latest
```
