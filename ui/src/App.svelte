<svelte:head>
    <title>WireGuard VPN</title>
</svelte:head>

<script>
    import About from "./About.svelte";
    import Clients from "./Clients.svelte";
    import EditClient from "./EditClient.svelte";
    import Nav from "./Nav.svelte";
    import NewClient from "./NewClient.svelte";

    import Cookie from "cookie-universal";
    import {Route, Router} from "svelte-routing";

    const cookies = Cookie();
    export let user = cookies.get("wguser", {fromRes: true}) || "anonymous";

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
                <Route path="client/:clientId" component="{EditClient}"/>
                <Route path="newclient/" component="{NewClient}"/>
                <Route path="about" component="{About}"/>
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
