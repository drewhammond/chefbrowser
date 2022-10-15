import * as React from 'react';
import Head from 'next/head';
import Navbar from "../components/navbar";

export default function MyApp(props) {
	const {Component, pageProps} = props;
	return (
		<>
			<Head>
				<meta name="viewport" content="initial-scale=1, width=device-width"/>
			</Head>
				<Navbar/>
				<div id="main" className="container">
					<Component {...pageProps} />
				</div>
		</>
	);
}
