<script lang="ts">
    import Fab, {Icon, Label} from "@smui/fab";
    import Dialog, {Actions, InitialFocus} from "@smui/dialog";
    import Textfield from "@smui/textfield";
    import HelperText from "@smui/textfield/helper-text";
    import Button from "@smui/button";
    import {Content, Title} from "@smui/paper";
    import api, {type WGClient} from "./lib/api";

    import Cookie from "cookie-universal";
    import {onMount} from "svelte";
    import {navigate} from "svelte-routing";
    import {convertNETIPToTextCIDRs, convertTextCIDRsToNETIP} from "./lib/ip";
    import {parseJwt} from "./lib/jwt";

    export let clientId: string;

    const {user} = parseJwt<{user: string}>(Cookie().get("wguser", {fromRes: true}));

    let client: WGClient = {
        allowedIps: [],
        name: "",
        notes: "",
        created: "",
        updated: "",
        mtu: 0,
        ip: "",
        publicKey: "",
        psk: "",
    };
    let clientName = "";
    let clientNotes = "";
    let allowedIPsText = "";
    let openDeleteDialog: boolean = false;

    async function getClient() {
        const c = await api.get(user, decodeURIComponent(clientId));
        clientName = c.name;
        clientNotes = c.notes;
        client = c;
        allowedIPsText = (c.allowedIps || []).join("\n");
        console.log("Fetched client", client);
    }

    async function handleSubmit(event: Event) {
        client.name = clientName;
        client.notes = clientNotes;
        const CIDR_REGEX = /^(\d{1,3}\.){3}\d{1,3}(\/\d{1,2})?$/;
        client.allowedIps = allowedIPsText.split("\n").filter((cidr) => {
            cidr = cidr.trim();
            return CIDR_REGEX.test(cidr);
        });
        client = await api.update(user, client.publicKey, client);
        navigate("/", {replace: true});
        console.log("Saved changes", client);
    }

    async function deleteHandler(e: any) {
        switch (e.detail.action) {
            case "delete":
                await api.delete(user, client.publicKey);
                navigate("/", {replace: true});
                break;
            default:
                break;
        }
    }

    onMount(getClient);
</script>

<div class="back">
    <Fab color="primary" href="/">
        <Icon class="material-icons">arrow_back</Icon>
    </Fab>
</div>

<h3 class="mdc-typography--headline3">Client Properties <small class="text-muted">({client.name})</small></h3>

<div class="container">
    <form on:submit|preventDefault={handleSubmit}>
        <div class="margins">
            <Textfield
                input$id="name"
                bind:value={clientName}
                variant="outlined"
                label="Client Name"
                aria-controls="client-name"
                aria-describedby="client-name-help"
                style="width: 100%;"
                helperLine$style="width: 100%;"
            >
                <HelperText id="client-name-help" slot="helper">Friendly name of client / device</HelperText>
            </Textfield>
        </div>

        <div class="margins">
            <Textfield
                input$id="notes"
                textarea
                bind:value={clientNotes}
                label="Label"
                aria-controls="client-notes"
                style="width: 100%;"
                helperLine$style="width: 100%;"
                aria-describedby="client-notes-help"
            >
                <HelperText slot="helper" id="client-notes-help">Notes about the client.</HelperText>
            </Textfield>
        </div>
        <div class="margins">
            <Textfield
                input$id="allowedIps"
                style="width: 100%;"
                helperLine$style="width: 100%;"
                textarea
                bind:value={allowedIPsText}
                label="Allowed IPs"
            >
                <HelperText id="client-notes-help" slot="helper"
                    >Additional allowed CIDR blocks accessible via the client separated by a newline
                </HelperText>
            </Textfield>
        </div>

        <Button variant="raised"><Label>Save Changes</Label></Button>
    </form>
</div>

<div class="container">
    <h3 class="mdc-typography--headline5">Additional Properties</h3>
    <dl>
        <dt>IP Address</dt>
        <dd>{client.ip}</dd>
        <dt>Private Key</dt>
        <dd>{client.privateKey}</dd>
        <dt>Public key</dt>
        <dd>{client.publicKey}</dd>
        <dt>Preshared Key</dt>
        <dd>{client.psk}</dd>
    </dl>
</div>

<div class="container">
    <h3 class="mdc-typography--headline4">Danger Zone</h3>

    <Dialog
        open={openDeleteDialog}
        aria-labelledby="delete-title"
        aria-describedby="delete-content"
        on:MDCDialog:closed={deleteHandler}
    >
        <div class="container">
            <Title id="delete-title">Delete Client Config</Title>
            <Content id="delete-content">Are you sure you want to delete this client configuration?</Content>
            <Actions>
                <Button action="none">
                    <Label>No</Label>
                </Button>
                <Button action="delete" use={[InitialFocus]}>
                    <Label>Yes</Label>
                </Button>
            </Actions>
        </div>
    </Dialog>

    <Button id="delete" variant="raised" on:click={() => (openDeleteDialog = true)}
        ><Label>Delete Client Config</Label>
    </Button>
</div>

<style>
    .back {
        position: fixed;
        left: 10px;
        top: 70px;
    }
</style>
