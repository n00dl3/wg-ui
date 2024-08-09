<script lang="ts">
    import Button from "@smui/button";
    import Textfield from "@smui/textfield";
    import Cipher from "./lib/master-key";
    import {CipherException} from "./lib/master-key.js";

    let passphrase = "";
    let encryptionKey = "";

    function handleSubmit(e:Event) {
        e.preventDefault();
        if (Cipher.wrappedKey) {
            Cipher.unlock(window.location.hash.substring(1), passphrase);
        } else {
            Cipher.init(passphrase);
        }
        Cipher.unlock(window.location.hash.substring(1), passphrase).then((key) => {
            
        });
        
    }
</script>

<div class="master-key-unlock">
    <form on:submit={handleSubmit}>
    {#if Cipher.wrappedKey}
        <p>Enter your passphrase to unlock the master key</p>
    {:else}
        <p>Master key is not set, a new key will be generated, please provide a passphrase for the newly generated key</p>
    {/if}
        <Textfield
            label="Passphrase"
            type="password"
            bind:value={passphrase}
        />
        <Button type="submit">{Cipher.wrappedKey?'Unlock Key':'Generate key'}</Button>
    </form>
</div>

<style>
    .master-key-unlock {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 16px;
    }

    p {
        margin-top: 16px;
    }
</style>