<script lang="ts">
    import Button from "@smui/button";
    import Textfield from "@smui/textfield";
    import Snackbar, {Actions, Label} from "@smui/snackbar";
    import IconButton from "@smui/icon-button";
    import {createEventDispatcher} from "svelte";
    import Cipher from "./lib/master-key";
    import api from "./lib/api";
    export let user: string;
    let passphrase = "";
    let encryptionKey = "";
    let showWarning = false;
    let validPassphrase = true;
    let errorSnackbar: Snackbar;
    const wrappedKey = window.location.hash.substring(1);
    const dispatch = createEventDispatcher();
    api.list(user).then((clients) => {
        if (clients.length > 0 && !wrappedKey) {
            showWarning = true;
        } else {
            showWarning = false;
        }
    });

    function handleSubmit(e: Event) {
        e.preventDefault();
        if (wrappedKey) {
            Cipher.unlock(wrappedKey, passphrase)
                .then((k) => {
                    validPassphrase = true;
                    dispatch("unlock", k);
                })
                .catch(() => {
                    validPassphrase = false;
                    errorSnackbar.open();
                });
        } else {
            Cipher.init(passphrase).then((k) => dispatch("unlock", k));
        }
    }
</script>

<div class="master-key-unlock">
    <form on:submit={handleSubmit}>
        {#if wrappedKey}
            <p>Enter your passphrase to unlock the master key</p>
        {:else}
            <p>
                Master key is not set, a new key will be generated, please provide a passphrase for the newly generated
                key
            </p>
        {/if}
        <Textfield label="Passphrase" type="password" bind:value={passphrase} invalid={!validPassphrase} />
        <Button type="submit">{wrappedKey ? "Unlock Key" : "Generate key"}</Button>
    </form>
    <Snackbar bind:this={errorSnackbar} class="demo-error">
        <Label>That thing you tried to do didn't work. Honestly, I'm not sure why you even tried.</Label>
        <Actions>
            <IconButton class="material-icons" title="Dismiss">close</IconButton>
        </Actions>
    </Snackbar>
</div>

<style lang="scss">
    // Make sure SMUI's import happens first, since it specifies variables.
    @use '@smui/snackbar/style.scss' as smui-snackbar;
    // See https://github.com/material-components/material-components-web/tree/v14.0.0/packages/mdc-snackbar
    @use '@material/snackbar/mixins' as snackbar;
    // See https://github.com/material-components/material-components-web/tree/v14.0.0/packages/mdc-theme
    @use '@material/theme/color-palette';
    @use '@material/theme/theme-color';
    .master-key-unlock {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 16px;
    }

    p {
        margin-top: 16px;
    }
    .mdc-snackbar.demo-error {
        @include snackbar.fill-color(color-palette.$red-500);
        @include snackbar.label-ink-color(theme-color.accessible-ink-color(color-palette.$red-500));
    }
</style>
