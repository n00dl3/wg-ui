<script lang="ts">
    import Button, {Label} from '@smui/button';
    import IconButton from '@smui/icon-button';
    import Paper from '@smui/paper';

    export let qrCodeURI: string;
    export let configURI: string;
    export let id: string;
    export let name: string;
    export let ip: string;
    export let publicKey: string;
    export let privateKey: string;

    let hash = 0;
    for (let i = 0; i < privateKey.length; i++) {
        hash = privateKey.charCodeAt(i) + ((hash << 5) - hash);
    }
    const color = `hsl(${(hash % 360)},50%,95%)` ;

</script>

<style>
    @media screen and (max-width: 800px) {
        img {
            display: none;
        }
    }

    img {
        margin-right: 40px;
        border: 1px solid #ccc;
    }

    .download {
        margin-top: 2em;
    }
</style>

<Paper elevation="{8}" style="background-color: {color}; margin: 2em 0;" class="card">

    <div class="float-right">
        <IconButton class="float-right material-icons" href="/client/{id}">edit</IconButton>
    </div>


    <img src="{qrCodeURI}" class="qrcode float-right" alt="Mobile client config"/>

    <i class="material-icons" aria-hidden="true">devices</i>
    <h3 class="mdc-typography--headline5">{name}</h3>

    <dl>
        <dt>IP</dt>
        <dd>{ip}</dd>
        <dt>Public Key</dt>
        <dd>{publicKey}</dd>
    </dl>

    <div class="download">
        <Button href="{configURI}" variant="raised"><Label>Download
            Config</Label></Button>
    </div>
</Paper>
