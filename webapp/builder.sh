#!/bin/sh
set -u

publish() {
    release="$SITE_DIR/release-$(date +%s)"
    cp -r build "$release" || return 1
    ln -sfn "$release" "$SITE_DIR/next" || return 1
    mv -T "$SITE_DIR/next" "$SITE_DIR/current" || return 1
    for old in "$SITE_DIR"/release-*; do
        [ "$old" != "$release" ] && rm -rf "$old"
    done
    return 0
}

while true; do
    if npm run build && publish; then
        echo "site published"
        sleep "$WEBAPP_INTERVAL"
    else
        echo "build failed, keeping the published site; retrying in ${WEBAPP_RETRY}s"
        sleep "$WEBAPP_RETRY"
    fi
done
