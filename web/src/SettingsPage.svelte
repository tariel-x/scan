<script>
    import {
        Col,
        Row,
    } from 'sveltestrap';

    export let active;
    $: mainRowClass = active ? '' : 'd-none';


    const clear = () => {
    };

    import { scanner } from './stores.js';
    let scannerName;
    const unsubscribe = scanner.subscribe(value => {
        scannerName = value;
        clear();
    });


</script>

<Row class={mainRowClass}>
    <Col>
        Settings
    </Col>
</Row>