<script lang="ts">
    import Fab, {Icon, Label} from "@smui/fab";
    import Textfield from "@smui/textfield";
    import HelperText from "@smui/textfield/helper-text";
    import Button from "@smui/button";
    import Switch from "@smui/switch";
    import FormField from "@smui/form-field";
    import Cookie from "cookie-universal";
    import { navigate} from "svelte-routing";
    import api, {type WGClientCreateForm} from "./lib/api.js";
    import {parseJwt} from "./lib/jwt.js";
    import {generatePresharedKey, generatePrivateKey, generatePublicKey, keyToHex} from "./lib/keygen.js";
    import Dialog, {Actions, Content, Header, Title} from "@smui/dialog";
    import {generateConfig, getQrcodeConfig} from "./lib/config.js";
    import QRCode from "qrcode";
    import Cipher from "./lib/master-key.js";

    let qrCodeUri: string;
    let showDialog = false;
    let configuration: string;
    let downloadUri: string;
    const {user} = parseJwt<{user: string}>(Cookie().get("wguser", {fromRes: true}));

    let client = {
        generatePSK: false,
        Name: "",
        Notes: "",
    };

    const handleSubmit = async () => {
        const privateKey = generatePrivateKey();
        const publicKey = generatePublicKey(privateKey);
        const psk = client.generatePSK ? generatePresharedKey() : undefined;
        const c: WGClientCreateForm = {
            name: client.Name,
            notes: client.Notes,
            publicKey: keyToHex(publicKey),
            privateKey:keyToHex(await Cipher.encrypt(privateKey)),
            psk: psk?.length ? keyToHex(psk) : undefined,
            allowedIPs: [],
        };
        const result = await api.create(user, c);
        configuration = generateConfig({
            dns: result.dns,
            ipAddress: result.ip,
            keepAlive: result.keepAlive,
            mtu: result.mtu,
            name: result.name,
            notes: result.notes,
            privateKey: keyToHex(privateKey),
            serverAllowedIps: result.allowedIps || [],
            serverEndpoint: result.server.endpoint,
            serverPublicKey: result.server.publicKey,
            presharedKey: psk ? keyToHex(psk) : undefined,
        });

        qrCodeUri = await QRCode.toDataURL(configuration);
        downloadUri = "data:text/plain;charset=utf-8," + encodeURIComponent(configuration);
        showDialog = true;
    };
</script>

<div class="back">
    <Fab on:click$preventDefault={() => navigate(`/${window.location.hash}`)} href="/{window.location.hash}" color="primary">
        <Icon class="material-icons">arrow_back</Icon>
    </Fab>
</div>

<h3 class="mdc-typography--headline3"><small>Create New Device Configuration</small></h3>

<div class="container">
    <form on:submit|preventDefault={handleSubmit}>
        <div class="margins">
            <Textfield
                input$id="name"
                style="width: 100%;"
                helperLine$style="width: 100%;"
                bind:value={client.Name}
                variant="outlined"
                label="Client Name"
                aria-controls="client-name"
                aria-describedby="client-name-help"
            >
                <HelperText slot="helper" id="client-name-help">Friendly name of client / device</HelperText>
            </Textfield>
        </div>

        <div class="margins">
            <Textfield
                input$id="notes"
                style="width: 100%;"
                helperLine$style="width: 100%;"
                textarea
                bind:value={client.Notes}
                label="Label"
                aria-controls="client-notes"
                aria-describedby="client-notes-help"
            >
                <HelperText slot="helper" id="client-notes-help">Notes about the client.</HelperText>
            </Textfield>
        </div>
        <div class="margins">
            <FormField style="margin-bottom: 2em;">
                <Switch bind:checked={client.generatePSK} />
                <span slot="label">Generate a Pre-shared Key</span>
            </FormField>
        </div>

        <Button variant="raised"><Label>Create</Label></Button>
    </form>
    <Dialog open={showDialog} on:SMUIDialog:closed={() => navigate(`/${window.location.hash}`, {replace: true})}>
        <Header>
            <Title>Device created</Title>
        </Header>
        <Content>
            <div style="text-align: center;">
                <p>
                    Your new device has been added, scan the QRcode or click on the "DOWNLOAD" button below to retrieve
                    your device configuration
                </p>
                <img src={qrCodeUri} alt="Scan this QRCode to get your new client configuration" />
                <p>
                    Keep this configuration file carefully, you won't be able to retrieve it once this dialog is closed
                </p>
            </div>
        </Content>
        <Actions>
            <Button action="close">
                <Label>Close</Label>
            </Button>
            <Button href={downloadUri} download="config.cfg" variant="raised" tag="a">
                <Label>Download Config</Label>
            </Button>
        </Actions>
    </Dialog>
</div>

<style>
    .back {
        position: fixed;
        left: 10px;
        top: 70px;
    }
</style>
