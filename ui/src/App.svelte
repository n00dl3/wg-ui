<script lang="ts">
    import About from "./About.svelte";
    import Clients from "./Clients.svelte";
    import EditClient from "./EditClient.svelte";
    import {parseJwt} from "./lib/jwt";
    import Cipher from "./lib/master-key";
    import Nav from "./Nav.svelte";
    import NewClient from "./NewClient.svelte";

    import Cookie from "cookie-universal";
    import {Route, Router} from "svelte-routing";
    import MasterkeyInput from "./MasterkeyInput.svelte";
    import {keyToHex} from "./lib/keygen";

    const cookie = Cookie().get("wguser", {fromRes: true});
    let unlocked = Cipher.isUnlocked();
    const handleUnlock = () => {
        unlocked = Cipher.isUnlocked();
        if (unlocked) {
            window.location.hash = Cipher.wrappedKey ? "#" + keyToHex(Cipher.wrappedKey) : "#";
        }
    };
    export let user: string;
    try {
        const token = parseJwt<{user: string}>(cookie || "");
        user = token.user;
    } catch (e) {
        user = "anonymous";
    }
    export let url = "";
</script>

<svelte:head>
    <title>WireGuard VPN</title>
</svelte:head>

<div class="mdc-typography">
    <Router {url}>
        <Nav {user} />
        {#if unlocked}
            <main class="container">
                <div>
                    <Route path="client/:clientId" component={EditClient} />
                    <Route path="newclient/" component={NewClient} />
                    <Route path="about" component={About} />
                    <Route path="/">
                        <Clients {user} />
                    </Route>
                </div>
            </main>
        {:else}
            <MasterkeyInput on:unlock={handleUnlock} />
        {/if}
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

<style>
    footer {
        margin-top: 3em;
        border-top: 1px solid #ddd;
        text-align: center;
    }
</style>
