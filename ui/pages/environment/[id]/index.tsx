import {useRouter} from 'next/router'
import Link from 'next/link'
import useSWR from "swr";

// @ts-ignore
const fetcher = (...args) => fetch(...args).then(res => res.json())

export default function EnvironmentPage() {
	const router = useRouter()
	const id = router.query.id as string

	const { data, error } = useSWR(process.env.BASE_URL + `/api/environment/${id}`, fetcher)

	if (error) return <div>failed to load</div>
	if (!data) return <div>loading...</div>

	return (
		<>
			<h1>Environment: {id}</h1>

		</>
	)
}
