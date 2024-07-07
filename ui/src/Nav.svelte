<script>
    import TopAppBar, {Row, Section, Title} from "@smui/top-app-bar";

    export let user;
</script>

<style>
    @media screen and (max-width: 800px) {
        .user {
            display: none;
        }
    }
</style>

<TopAppBar variant="static" color="primary">
    <Row>
        <Section>
            <Title>WireGuard VPN</Title>
        </Section>
        <Section align="end" toolbar>
            <small class="user">Logged in as {user}</small>
        </Section>
    </Row>
</TopAppBar>
