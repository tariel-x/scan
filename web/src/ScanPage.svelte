<script>
    import {
        Col,
        Row,
        Button,
        Media,
        ButtonGroup,
        Spinner,
    } from 'sveltestrap';

    export let settings = {};

    export let active;
    $: mainRowClass = active ? '' : 'd-none';

    let imageSource = '';
    let imageShown = false;
    $: imageClass = imageShown ? '' : 'invisible';

    let spinnerShown = false;
    $: spinnerClass = spinnerShown ? '' : 'invisible';

    const clear = () => {
        imageSource = "";
        imageShown = false;
    };

    import { scanner } from './stores.js';
    let scannerName;
    const unsubscribe = scanner.subscribe(value => {
        scannerName = value;
        clear();
    });

    const scan = async () => {
        spinnerShown = true;
        await fetch(`/api/devices/`+ scannerName +`/scan`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(settings)
        })
            .then(r => r.blob())
            .then(blob => {
                spinnerShown = false;
                imageSource = (window.URL || window.webkitURL).createObjectURL(blob);
                imageShown = true;
            });
    };

    const download = () => {
        var a = document.createElement('a');
        if (window.URL && window.Blob && ('download' in a) && window.atob) {
            a.href = imageSource;
            a.download = "scanned.png";
            a.click();
            window.URL.revokeObjectURL(imageSource);
        }
    };

</script>

<Row class={mainRowClass}>
    <Col>
        <Row>
            <Col>
                <ButtonGroup>
                    <Button outline primary on:click={scan}>Scan</Button>
                    <Button outline success disabled={!imageShown} on:click={download}>Download</Button>
                    <Button outline on:click={clear}>Clear</Button>
                </ButtonGroup>
            </Col>
        </Row>
        <Row>
            <Col>
                <h5 object class={imageClass}>Image from {scannerName}</h5>
                <img class="img-fluid {imageClass}" src={imageSource}/>
                <Spinner class={spinnerClass}/>
            </Col>
        </Row>
    </Col>
</Row>