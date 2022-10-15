import Link from "next/link";
import * as React from "react";

export default function Navbar() {
	return (
		<nav className="navbar navbar-expand-lg bg-light">
			<div className="container">
				<button className="navbar-toggler" type="button" data-bs-toggle="collapse"
						data-bs-target="#navbarSupportedContent" aria-controls="navbarSupportedContent"
						aria-expanded="false" aria-label="Toggle navigation">
					<span className="navbar-toggler-icon"></span>
				</button>
				<Link href={`/nodes`}>
					<a className="navbar-brand">Chef Browser</a>
				</Link>
				<div className="collapse navbar-collapse" id="navbarNavAltMarkup">
					<div className="navbar-nav">
						<Link href={`/nodes`}>
							<a className="nav-link">Nodes</a>
						</Link>
						<Link href={`/environments`}>
							<a className="nav-link">Environments</a>
						</Link>
						<Link href={`/roles`}>
							<a className="nav-link">Roles</a>
						</Link>
						<Link href={`/databags`}>
							<a className="nav-link">Databags</a>
						</Link>
						<Link href={`/cookbooks`}>
							<a className="nav-link">Cookbooks</a>
						</Link>
					</div>
				</div>
			</div>
		</nav>
	)
}
