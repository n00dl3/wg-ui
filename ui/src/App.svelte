<svelte:head>
    <title>WireGuard VPN</title>
</svelte:head>

<script lang="ts">
    import About from "./About.svelte";
    import Clients from "./Clients.svelte";
    import EditClient from "./EditClient.svelte";
    import { parseJwt } from "./lib/jwt";
    import Cipher from "./lib/master-key";
    import Nav from "./Nav.svelte";
    import NewClient from "./NewClient.svelte";

    import Cookie from "cookie-universal";
    import {Route, Router} from "svelte-routing";
    const cookie = Cookie().get("wguser", {fromRes: true});
    if(window.location.hash){
        const password=prompt("Please provide the password to unlock the master key", "")||"";
        Cipher.unlock(window.location.hash.substring(1),password).then(()=>console.log("key unlocked !"))
    }else{
        const password=prompt("A new master key will be generated, please provide a password to encrypt it", "");
        if(!password){
            throw new Error("No password provided");
        }
        Cipher.init(password).then((key)=>{
            window.location.hash=key;
        });

    }
    export let user:string;
    if(cookie){
        const token = parseJwt<{user:string}>(cookie);
        user = token.user;
    }else{
        user="anonymous";
    }
    export let url = "";
</script>

<style>

    footer {
        margin-top: 3em;
        border-top: 1px solid #ddd;
        text-align: center;

    }
</style>

<div class="mdc-typography">

    <Router url="{url}">

        <Nav user="{user}"/>

        <main class="container">
            <div>
                <Route path="client/:clientId" component="{EditClient}" />
                <Route path="newclient/" component="{NewClient}"/>
                <Route path="about" component="{About}" />
                <Route path="/">
                    <Clients user="{user}"/>
                </Route>
            </div>
        </main>

    </Router>

    <footer>
        <p>
            Powered by <a href="https://github.com/EmbarkStudios/wg-ui">WG UI</a>.
        </p>
        <p>
            Copyright &copy; 2021 <a href="https://embark-studios.com">Embark Studios</a>.
        </p>
    </footer>
</div>
