<script lang="ts">
    import Button, {Icon, Label} from '@smui/button';
    import Paper from '@smui/paper';

    export let name: string;
    export let ip: string;
    export let publicKey: string;

    let hash = 0;
    for (let i = 0; i < publicKey.length; i++) {
        hash = publicKey.charCodeAt(i) + ((hash << 5) - hash);
    }
    const color = `hsl(${(hash % 360)},50%,95%)` ;
</script>

<style>
    .download {
        margin-top: 2em;
        text-align: right;
    }
</style>

<Paper elevation="{8}" style="background-color: {color}; margin: 2em 0;" class="card">



    <i class="material-icons" aria-hidden="true">devices</i>
    <h3 class="mdc-typography--headline5">{name}</h3>

    <dl>
        <dt>IP</dt>
        <dd>{ip}</dd>
        <dt>Public Key</dt>
        <dd>{publicKey}</dd>
    </dl>

    <div class="download">
        <Button href="/client/{publicKey}" variant="raised"><Label><Icon class="material-icons">edit</Icon>Edit</Label></Button>
    </div>
</Paper>
