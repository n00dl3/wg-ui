<script lang="ts">
    import Button, {Icon, Label} from "@smui/button";
    import Paper from "@smui/paper";
    import {link, navigate} from "svelte-routing";
    import {generateConfig} from "./lib/config";
    import {keyFromHex, keyToBase64, keyToHex} from "./lib/keygen";
    import type {WGClient} from "./lib/api";
    import QRCode from "qrcode";
    import Cipher from "./lib/master-key";
    export let peerConfig: WGClient;


    let hash = 0;
    for (let i = 0; i < peerConfig.publicKey.length; i++) {
        hash = peerConfig.publicKey.charCodeAt(i) + ((hash << 5) - hash);
    }
    const color = `hsl(${hash % 360},50%,95%)`;
    let privateKey = "";
    Cipher.decrypt(keyFromHex(peerConfig.privateKey)).then((key) => {
        privateKey = keyToBase64(key);
        configuration = generateConfig({
            dns: peerConfig.dns,
            ipAddress: peerConfig.ip,
            keepAlive: peerConfig.keepAlive,
            mtu: peerConfig.mtu,
            name: peerConfig.name,
            notes: peerConfig.notes,
            privateKey: privateKey,
            serverAllowedIps: peerConfig.allowedIps || [],
            serverEndpoint: peerConfig.server.endpoint,
            serverPublicKey: peerConfig.server.publicKey,
            presharedKey: peerConfig.psk ? keyToBase64(keyFromHex(peerConfig.psk)) : undefined,
        });
        QRCode.toDataURL(configuration).then((uri) => {
            qrCodeUri = uri;
        });
        downloadUri = "data:text/plain;charset=utf-8," + encodeURIComponent(configuration)
    });
    let configuration: string = "";
    let qrCodeUri = "";

    let downloadUri = "";

</script>

<Paper elevation={8} style="background-color: {color}; margin: 2em 0;" class="card">
    <img
        src="{qrCodeUri}"
        class="qrcode float-right"
        alt="Mobile client config"
    />
    <i class="material-icons" aria-hidden="true">devices</i>
    <h3 class="mdc-typography--headline5">{name}</h3>

    <dl>
        <dt>IP</dt>
        <dd>{peerConfig.ip}</dd>
        <dt>Public Key</dt>
        <dd>{peerConfig.publicKey}</dd>
    </dl>

    <div class="download">
        <Button
            tag="a"
            on:click$preventDefault={() => navigate(`/client/${peerConfig.publicKey}${window.location.hash}`)}
            href="/client/{peerConfig.publicKey}{window.location.hash}"
            variant="raised"><Label><Icon class="material-icons">edit</Icon>Edit</Label></Button
        >
    </div>
</Paper>

<style>
    .download {
        margin-top: 2em;
        text-align: right;
    }
</style>
