<script lang="ts">
    import Button from "@smui/button";
    import Textfield from "@smui/textfield";
    import Cipher from "./lib/master-key";

    let passphrase = "";
    let encryptionKey = "";

    function handleSubmit(e:Event) {
        e.preventDefault();

        Cipher.unlock(window.location.hash.substring(1), passphrase).then((key) => {
            
        });
        
    }
</script>

<main>
    {#if Cipher.wrappedKey}
    <form on:submit={handleSubmit}>
        <p>Enter your passphrase to unlock the master key</p>
        <Textfield
        label="Passphrase"
        type="password"
        value={passphrase}
        on:input={(event) => (passphrase = event?.target.value)}
        />
        
        <Button type="submit">Unlock Key</Button>
        
    </form>
    {:else}
    <p>Master key is not set</p>
    {/if}
    
</main>

<style>
    main {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 16px;
    }

    p {
        margin-top: 16px;
    }
</style>