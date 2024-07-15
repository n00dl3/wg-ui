<script lang="ts">
    import Fab, {Icon} from "@smui/fab";
    import List, {Item, Graphic, PrimaryText, SecondaryText, Meta, Text} from "@smui/list";
    import {onMount} from "svelte";
    import Client from "./Client.svelte";
    import api, {type WGClient} from "./lib/api";

    export let user: string;
    let clients: [string, WGClient][] = [];

    async function getClients() {
        clients = await api.list(user);
    }

    onMount(getClients);
</script>

<div class="content">
    <div class="row">
        <div class="col">
            <h2 class="mdc-typography--headline2">
                My VPN Clients<small class="mdc-typography--headline5"
                    >({user}
                    )</small
                >
            </h2>
        </div>
        <div class="col help">
            <h3>Instructions</h3>
            <ol>
                <li><a href="https://www.wireguard.com/install/">Install WireGuard</a></li>
                <li>Download your WireGuard config</li>
                <li>Connect to the VPN server</li>
            </ol>
        </div>
    </div>
</div>

{#each clients as [id, dev]}
    <Client {user} {id} ip={dev.ip} name={dev.name} publicKey={dev.publicKey} />
{/each}

<div class="newClient">
    <Fab color="primary" href="/newclient">
        <Icon class="material-icons">add</Icon>
    </Fab>
</div>

<style>
    .newClient {
        float: right;
    }

    h2 small {
        display: block;
        clear: left;
        color: #ccc;
    }

    .row {
        display: flex;
        flex-direction: row;
        flex-wrap: wrap;
        width: 100%;
    }

    .col {
        display: flex;
        flex-direction: column;
        flex-basis: 100%;
        flex: 1;
        margin-left: 2em;
    }

    .help {
        flex-basis: 10%;
    }

    h2 {
        margin: 0;
        padding: 0;
    }
</style>
