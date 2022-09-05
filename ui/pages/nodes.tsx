import * as React from "react";
import useSWR from 'swr'
import Link from "next/link";

// @ts-ignore
const fetcher = (...args) => fetch(...args).then(res => res.json())

export default function NodesPage() {
	const {data, error} = useSWR(process.env.BASE_URL + '/api/nodes', fetcher)

	if (error) return <div>failed to load</div>
	if (!data) return <div>loading...</div>

	return (
		<>
			<ul className="list-unstyled">
				{data.nodes.map(node => {
					return (
						<Link href={`/node/${node}`}>{node}</Link>
					)
				})}
			</ul>
		</>
	)
}
