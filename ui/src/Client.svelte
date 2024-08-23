<script lang="ts">
    import Button, { Icon, Label } from "@smui/button";
    import LayoutGrid, { Cell } from "@smui/layout-grid";
    import Paper from "@smui/paper";
    import QRCode from "qrcode";
    import { navigate } from "svelte-routing";
    import type { WGClient } from "./lib/api";
    import { generateConfig, type PeerConfig } from "./lib/config";
    import { keyFromHex, keyToBase64 } from "./lib/keygen";
    import Cipher from "./lib/master-key";
    export let peerConfig: WGClient;
    let decipherError = false;
    const config:PeerConfig = {
        dns: peerConfig.dns,
        ipAddress: peerConfig.ip,
        keepAlive: peerConfig.keepAlive,
        mtu: peerConfig.mtu,
        name: peerConfig.name,
        notes: peerConfig.notes,
        privateKey: "",
        serverAllowedIps: peerConfig.server.allowedIPs || [],
        serverEndpoint: peerConfig.server.endpoint,
        serverPublicKey: peerConfig.server.publicKey,
        presharedKey: peerConfig.psk ? keyToBase64(keyFromHex(peerConfig.psk)) : undefined,
    };
    let privateKey = "";
    let configuration: string = "";
    let qrCodeUri = "";
    let downloadUri = "";
    let hash = 0;
    for (let i = 0; i < peerConfig.publicKey.length; i++) {
        hash = peerConfig.publicKey.charCodeAt(i) + ((hash << 5) - hash);
    }
    const color = `hsl(${hash % 360},50%,95%)`;
    Cipher.decrypt(keyFromHex(peerConfig.privateKey))
        .then((key) => {
            privateKey = keyToBase64(key);
            config.privateKey = privateKey;
        })
        .catch((e) => {
            console.error(e);
            decipherError = true;
        }).finally(()=>{
            configuration = generateConfig(config);
            QRCode.toDataURL(configuration).then((uri) => {
                qrCodeUri = uri;
            });
            downloadUri = "data:text/plain;charset=utf-8," + encodeURIComponent(configuration);
        });
    
</script>

<Paper elevation={8} style="background-color: {color}; margin: 2em 0;" class="card">
    <LayoutGrid>
        <Cell span={9}>
            <LayoutGrid>
                <Cell span={12}>
                    <h3 class="mdc-typography--headline5">
                        <i class="material-icons" aria-hidden="true">devices</i>&nbsp;{peerConfig.name}
                    </h3>
                    <dl>
                        <dt>IP</dt>
                        <dd>{peerConfig.ip}</dd>
                        <dt>Public Key</dt>
                        <dd>{peerConfig.publicKey}</dd>
                    </dl>
                </Cell>
                <Cell span={12}>
                    <div class="download">
                        <Button
                            tag="a"
                            on:click$preventDefault={() =>
                                navigate(`/client/${peerConfig.publicKey}${window.location.hash}`)}
                            href="/client/{peerConfig.publicKey}{window.location.hash}"
                            variant="raised"><Label><Icon class="material-icons">edit</Icon>Edit</Label></Button
                        >
                        <Button tag="a" href={downloadUri} download="client.cfg" variant="raised"
                            ><Label>Download Config</Label></Button
                        >
                    </div>
                </Cell>
            </LayoutGrid>
        </Cell>
        <Cell span={3}>
            <img src={qrCodeUri} class="qrcode float-right" alt="Mobile client config" />
        </Cell>
    </LayoutGrid>
</Paper>

<style>
    .download {
        text-align: right;
    }
    @media screen and (max-width: 800px) {
        img {
            display: none;
        }
    }

    img {
        margin: 1em;
        width: calc(100% - 2em);
    }
</style>
