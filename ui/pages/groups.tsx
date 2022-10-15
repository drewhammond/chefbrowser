import * as React from "react";
import useSWR from 'swr'
import Link from "next/link";

// @ts-ignore
const fetcher = (...args) => fetch(...args).then(res => res.json())

export default function GroupsPage() {
	const {data, error} = useSWR(process.env.BASE_URL + '/api/groups', fetcher)

	if (error) return <div>failed to load</div>
	if (!data) return <div>loading...</div>

	return (
		<>
			<ul className="list">
				{Object.keys(data).map((env, url) => {
					return (
						<Link href={`/group/${env}`}>
							<li>
								<a className={"foo"}>{env}</a>
							</li>
						</Link>
					)
				})}
			</ul>
		</>
	)
}
