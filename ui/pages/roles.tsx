import * as React from "react";
import useSWR from 'swr'
import Link from "next/link";

// @ts-ignore
const fetcher = (...args) => fetch(...args).then(res => res.json())

export default function RolesPage() {
	const {data, error} = useSWR(process.env.BASE_URL + '/api/roles', fetcher)

	if (error) return <div>failed to load</div>
	if (!data) return <div>loading...</div>

	if (data.roles.length == 0) {
		return (
			"no roles on server"
		)
	}

	return (
		<>
			<ul className="list-unstyled">
				{data.roles.map(role => {
					return (
						<Link href={`/role/${role}`}>{role}</Link>
					)
				})}
			</ul>
		</>
	)
}
