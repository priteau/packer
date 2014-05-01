#!/usr/bin/env bats
#
# This tests the nimbus builder. The teardown function will
# delete any images with the text "packerbats" within the name.

load test_helper
fixtures builder-nimbus

# Required parameters
: ${NIMBUS_S3ID:?}
: ${NIMBUS_S3KEY:?}
: ${NIMBUS_CANONICALID:?}
: ${NIMBUS_CERT:?}
: ${NIMBUS_KEY:?}
: ${NIMBUS_CLOUD_CLIENT_PATH:?}
command -v ${NIMBUS_CLOUD_CLIENT_PATH}/bin/cloud-client.sh >/dev/null 2>&1 || {
    echo "'cloud-client.sh' must be installed" >&2
    exit 1
}

USER_VARS="-var s3id=${NIMBUS_S3ID}"
USER_VARS="${USER_VARS} -var s3key=${NIMBUS_S3KEY}"
USER_VARS="${USER_VARS} -var canonicalid=${NIMBUS_CANONICALID}"
USER_VARS="${USER_VARS} -var cert=${NIMBUS_CERT}"
USER_VARS="${USER_VARS} -var key=${NIMBUS_KEY}"
USER_VARS="${USER_VARS} -var cloud_client_path=${NIMBUS_CLOUD_CLIENT_PATH}"

# This tests if GCE has an image that contains the given parameter.
nimbus_has_image() {
    ${NIMBUS_CLOUD_CLIENT_PATH}/bin/cloud-client.sh --conf ${NIMBUS_CLOUD_CLIENT_PATH}/conf/clouds/"$1".conf --list \
        | grep "packerbats-$1" | wc -l
}

teardown() {
    for cloud in alamo hotel hotel-kvm; do
        ${NIMBUS_CLOUD_CLIENT_PATH}/bin/cloud-client.sh --conf ${NIMBUS_CLOUD_CLIENT_PATH}/conf/clouds/"$cloud".conf --list \
            | grep packerbats | awk '{print $2; }' | sed "s/'//g" \
            | xargs -n1 ${NIMBUS_CLOUD_CLIENT_PATH}/bin/cloud-client.sh --conf ${NIMBUS_CLOUD_CLIENT_PATH}/conf/clouds/"$cloud".conf --delete --name
    done
}

@test "nimbus: build alamo.json" {
    run packer build ${USER_VARS} $FIXTURE_ROOT/alamo.json
    [ "$status" -eq 0 ]
    [ "$(nimbus_has_image "alamo")" -eq 1 ]
}

@test "nimbus: build hotel.json" {
    run packer build ${USER_VARS} $FIXTURE_ROOT/hotel.json
    [ "$status" -eq 0 ]
    [ "$(nimbus_has_image "hotel")" -eq 1 ]
}

@test "nimbus: build hotel-kvm.json" {
    run packer build ${USER_VARS} $FIXTURE_ROOT/hotel-kvm.json
    [ "$status" -eq 0 ]
    [ "$(nimbus_has_image "hotel-kvm")" -eq 1 ]
}
