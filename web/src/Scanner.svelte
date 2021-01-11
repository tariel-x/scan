<script>
    import {
        Col,
        Container,
        Row,
        Nav,
        NavItem,
        NavLink
    } from 'sveltestrap';
    import Scanners from './Scanners.svelte';
    import ScanPage from './ScanPage.svelte'
    import SettingsPage from './SettingsPage.svelte'

    let scannerName;

    import { scanner } from './stores.js';
    const unsubscribe = scanner.subscribe(value => {
        scannerName = value;
    });

    export let activeTab = 'scan';

    const selectTab = name => () => activeTab = name;
</script>

<Container>
    <Row>
        <Col sm="6" md="2">
            <Scanners/>
        </Col>
        <Col sm="6" md="10">
            <Row>
                <Col>
                    <Nav tabs>
                        <NavItem>
                            <NavLink active={"scan" === activeTab} on:click={selectTab("scan")}>Scan</NavLink>
                        </NavItem>
                        <NavItem>
                            <NavLink active={"settings" === activeTab} on:click={selectTab("settings")}>Settings</NavLink>
                        </NavItem>
                    </Nav>
                </Col>
            </Row>
            <ScanPage active={"scan" === activeTab}/>
            <SettingsPage active={"settings" === activeTab}/>
        </Col>
    </Row>
</Container>