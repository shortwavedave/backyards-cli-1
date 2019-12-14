#!/bin/bash
CHARTS_DIR=${1:-assets/charts}
ASSETS_DIR="${2:-assets}"

BACKYARDS_CHART_VERSION="1.1.0-dev.4"
ISTIO_OPERATOR_CHART_VERSION="0.0.30"
CANARY_OPERATOR_CHART_VERSION="0.1.8"
BACKYARDS_DEMO_CHART_VERSION="0.1.4"
CERT_MANAGER_CHART_VERSION="v0.10.0"
CERT_MANAGER_CRDS="https://raw.githubusercontent.com/jetstack/cert-manager/release-0.10/deploy/manifests/00-crds.yaml"
KAFKA_OPERATOR_CHART_VERSION="0.2.10"
ZOOKEEPER_OPERATOR_CHART_VERSION="0.0.2"
mkdir -p ${CHARTS_DIR};

CHARTS=("https://kubernetes-charts.banzaicloud.com/charts/istio-operator-${ISTIO_OPERATOR_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/canary-operator-${CANARY_OPERATOR_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-${BACKYARDS_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-demo-${BACKYARDS_DEMO_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-demo-${BACKYARDS_DEMO_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-demo-${BACKYARDS_DEMO_CHART_VERSION}.tgz")
CHARTS+=("https://charts.jetstack.io/charts/cert-manager-${CERT_MANAGER_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/kafka-operator-${KAFKA_OPERATOR_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/zookeeper-operator-${ZOOKEEPER_OPERATOR_CHART_VERSION}.tgz")

for i in ${CHARTS[@]}; do
    curl -s "${i}" | tar -zxv --directory "${CHARTS_DIR}/" -f -
    retVal=$?
    if [ $retVal -ne 0 ]; then
        exit $retVal
    fi
done

curl -s "${CERT_MANAGER_CRDS}" -o "${ASSETS_DIR}/cert-manager/crds.yaml"
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi
