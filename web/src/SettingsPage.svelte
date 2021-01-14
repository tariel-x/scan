<script>
    import {
        Col,
        Row,
        Spinner,
        Form,
        FormGroup,
        Input,
        Label,
    } from 'sveltestrap';

    import {onMount} from "svelte";

    import { scanner } from './stores.js';

    export let active;
    $: mainRowClass = active ? '' : 'd-none';

    let spinnerShown = false;
    $: spinnerClass = spinnerShown ? '' : 'd-none';

    let options = [];
    export let settings = {};

    const clear = () => {
        options = [];
    };

    const load = async (data) => {
        clear();
        spinnerShown = true;
        await fetch(`/api/devices/`+ scannerName +`/options`, {
            method: 'GET',
            body: JSON.stringify(data)
        })
            .then(r => r.json())
            .then(data => {
                spinnerShown = false;
                options = data;
                console.log(options);

                options.forEach(function(option) {
                    settings[option.name] = option.default
                });
            });
    };

    let scannerName;
    const unsubscribe = scanner.subscribe(value => {
        scannerName = value;
        if (scannerName !== "") {
            load();
        }
    });

    onMount(async () => {
        if (scannerName !== "") {
            await load();
        }
    });

</script>

<Row class={mainRowClass}>
    <Col>
        <Spinner class={spinnerClass}/>
        <Form>
            {#each options as option}
                <FormGroup>
                    {#if option.type === 0}
                        <Label for={option.name} check>
                            <Input id={option.name} placeholder={option.description} type="checkbox" bind:checked={settings[option.name]} />
                            {option.title}
                        </Label>
                    {:else if (option.type === 1 || option.type === 2) && option.set === null}
                        <Label for={option.name}>{option.title}
                            {#if option.range !== null} (Range: {option.range.min}-{option.range.max}, step: {option.range.quant}) {/if}
                        </Label>
                        <Input type="number" name={option.name} id={option.name} bind:value={settings[option.name]}/>
                    {:else if option.type === 3 && option.set === null}
                        <Label for={option.name}>{option.title}</Label>
                        <Input name={option.name} id={option.name} bind:value={settings[option.name]}/>
                    {:else if option.set !== null}
                        <Label for={option.name}>{option.title}</Label>
                        <Input type="select" name={option.name} id={option.name} bind:value={settings[option.name]}>
                            {#each option.set as item}
                                <option value={item}>{item}</option>
                            {/each}
                        </Input>
                    {:else}
                        Undefined option
                    {/if}
                </FormGroup>
            {/each}
        </Form>
    </Col>
</Row>