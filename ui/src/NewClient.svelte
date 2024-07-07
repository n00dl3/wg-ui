<script lang="ts">
    import Fab, {Icon, Label} from '@smui/fab';
    import Textfield from '@smui/textfield';
    import HelperText from '@smui/textfield/helper-text';
    import Button from '@smui/button';
    import Switch from '@smui/switch';
    import FormField from '@smui/form-field'
    import Cookie from "cookie-universal";
    import {navigate} from "svelte-routing";
    import api, {type WGClientCreateForm} from "./lib/api.js"

    const user = Cookie().get("wguser", {fromRes: true});

    let client:WGClientCreateForm = {
        generatePSK: false,
        Name: "",
        Notes: ""
    };

    const handleSubmit = async () => {
        await api.create(user, client);
        navigate("/", {replace: true});
    };

</script>

<style>
    .back {
        position: fixed;
        left: 10px;
        top: 70px;
    }
</style>

<div class="back">
    <Fab color="primary" href="/">
        <Icon class="material-icons">arrow_back</Icon>
    </Fab>
</div>

<h3 class="mdc-typography--headline3"><small>Create New Device Configuration</small></h3>

<div class="container">
    <form on:submit|preventDefault={handleSubmit}>

        <div class="margins">
            <Textfield input$id="name" style="width: 100%;"
                       helperLine$style="width: 100%;" bind:value={client.Name} variant="outlined" label="Client Name"
                       aria-controls="client-name" aria-describedby="client-name-help">

                <HelperText slot="helper" id="client-name-help">Friendly name of client / device</HelperText>
            </Textfield>
        </div>

        <div class="margins">
            <Textfield input$id="notes" style="width: 100%;"
                       helperLine$style="width: 100%;" textarea bind:value={client.Notes} label="Label"
                       aria-controls="client-notes" aria-describedby="client-notes-help">
                <HelperText slot="helper" id="client-notes-help">Notes about the client.</HelperText>
            </Textfield>
        </div>
        <div class="margins">
            <FormField style="margin-bottom: 2em;">
                <Switch bind:checked={client.generatePSK}/>
                <span slot="label">Generate a Pre-shared Key</span>
            </FormField>
        </div>

        <Button variant="raised"><Label>Create</Label></Button>
    </form>
</div>

