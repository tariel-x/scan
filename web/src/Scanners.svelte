<script>
    import { ListGroup, ListGroupItem, Spinner } from 'sveltestrap';

    import { onMount } from "svelte";

    let scanners;

    onMount(async () => {
        await fetch(`/api/devices`)
            .then(r => r.json())
            .then(data => {
                scanners = data;
                if (scanners.length > 0) {
                    select(scanners[0].name);
                }
            });
    });

    import { scanner } from './stores.js';

    function select(name) {
        scanner.set(name)
        value = name
    }

    let value = '';

</script>

{#if scanners}
    <ListGroup>
    {#each scanners as scanner }
        <ListGroupItem active={scanner.name == value} action tag="button" on:click={select(scanner.name)}>{scanner.name}</ListGroupItem>
    {/each}
    </ListGroup>
{:else}
    <Spinner/>
{/if}