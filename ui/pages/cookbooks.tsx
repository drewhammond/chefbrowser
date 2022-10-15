import * as React from "react";
import useSWR from 'swr'
import Link from "next/link";

// @ts-ignore
const fetcher = (...args) => fetch(...args).then(res => res.json())

export default function CookbooksPage() {
	const {data, error} = useSWR(process.env.BASE_URL + '/api/cookbooks', fetcher)

	if (error) return <div>failed to load</div>
	if (!data) return <div>loading...</div>

	return (
		<>
			<ul className="list">
				{data.cookbooks.map(cookbook => {
					return (
						<>
							<li>
								<a>{cookbook.name}</a>
							</li>
							<ul className={"list-unstyled"}>
								{cookbook.versions.map(version => {
									return (
										<Link href={`/cookbook/${cookbook.name}/${version}`}>
											<a className="mx-1">
												{version}
											</a>
										</Link>
									)
								})}
							</ul>
						</>
					)
				})}
			</ul>
		</>
	)
}
