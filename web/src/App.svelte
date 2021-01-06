<script>
	import { Col, Container, Row } from 'sveltestrap';
	import {
		Nav,
		NavItem,
		NavLink
	} from 'sveltestrap';

	let scannerName;
	import Scanners from './Scanners.svelte';
</script>

<main>
	<Container>
		<Row>
			<Col sm="6" md="2">
				<Scanners bind:value={scannerName}/>
			</Col>
			<Col sm="6" md="10">
				<Nav tabs>
					<NavItem>
						<NavLink href="#" active>Link</NavLink>
					</NavItem>
					<NavItem>
						<NavLink href="#">Link</NavLink>
					</NavItem>
				</Nav>
			</Col>
		</Row>
	</Container>
</main>