<script>
    import { ListGroup, ListGroupItem } from 'sveltestrap';
    import { Spinner } from 'sveltestrap';

    import { onMount } from "svelte";

    let scanners;

    onMount(async () => {
        await fetch(`http://localhost:8085/api/devices`)
            .then(r => r.json())
            .then(data => {
                scanners = data;
            });
    });

    export let value = '';

    const select = name => () => value = name;

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