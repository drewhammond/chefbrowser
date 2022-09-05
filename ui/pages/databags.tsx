import * as React from "react";
import useSWR from 'swr'
import Link from "next/link";

// @ts-ignore
const fetcher = (...args) => fetch(...args).then(res => res.json())

export default function DatabagsPage() {
	const {data, error} = useSWR(process.env.BASE_URL + '/api/databags', fetcher)

	if (error) return <div>failed to load</div>
	if (!data) return <div>loading...</div>

	return (
		<>
			<ul className="list-unstyled">
				{data.databags.map(databag => {
					return (
						<Link href={`/databag/${databag}`}>{databag}</Link>
					)
				})}
			</ul>
		</>
	)
}
