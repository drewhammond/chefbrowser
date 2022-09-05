import {useRouter} from 'next/router'
import Link from 'next/link'
import useSWR from "swr";
import * as React from "react";
import {splitCookiesString} from "next/dist/server/web/utils";
// @ts-ignore
const fetcher = (...args) => fetch(...args).then(res => res.json())

function AttributeRow(key, value) {
	return (
		<>
			<tr>
				<td><span className="font-monospace">${key}</span></td>
				<td><span className="font-monospace">{value}</span></td>
			</tr>
		</>
	)
}

function flatten(data, c) {
	var result = {}
	for (var i in data) {
		if (typeof data[i] == 'object') Object.assign(result, flatten(data[i], c + '.' + i))
		else result[(c + '.' + i).replace(/^\./, "")] = data[i]
	}
	return result
}

// we handle run list items differently (roles vs recipes)
function RunListLink(props) {
	const regex = /^(recipe|role)\[(.+)\]$/

	let item: string;
	({item} = props);

	let results = item.match(regex)

	if (results[1] == "recipe") {
		let s = results[2].split("::")
		let cookbook = s[0]
		// @ts-ignore
		let recipe = s[1]

		return (
			<Link href={`/cookbook/${cookbook}/recipes/${recipe}`}>
				<a className={"font-monospace"}>{results[0]}</a>
			</Link>
		)
	}

	if (results[1] == "role") {
		return (
			<Link href={`/roles/${results[2]}`}>
				<a>{results[0]}</a>
			</Link>
		)
	}

	return (
		<>Failed to parse</>
	)
}

function AttrTable(props) {
	let level, attributes;
	({level, attributes} = props);
	let flattenedAttrs = flatten(attributes, "")
	return (
		<table className="table table-sm table-responsive">
			<tbody>
			{Object.keys(attributes).map((k, v) => {
				return (
					<tr>
						<td><span className="font-monospace">${k}</span></td>
						<td><span className="font-monospace">{flattenedAttrs[k]}</span></td>
					</tr>
				)
			})}
			</tbody>
		</table>
	)
}

export default function NodePage() {
	const router = useRouter()
	const id = router.query.id as string

	const {data, error} = useSWR(`${process.env.BASE_URL}/api/node/${id}`, fetcher)

	if (error) return <div>failed to load</div>
	if (!data) return <div>loading...</div>

	return (
		<>
			<h2 className="font-monospace">{id}</h2>
			<hr/>
			{id} (ip)

			<ul className="list-unstyled">
				<li>
					<strong>Environment: </strong>
					<Link href={`/environment/${data.chef_environment}`}>
						<a>{data.chef_environment}</a>
					</Link>
				</li>
				<li>
					<strong>Run list:</strong> {data.run_list.map(item => {
					return (
						<RunListLink item={item}/>
					)
				})}
				</li>
			</ul>
			<br/>

			<h3>Attributes</h3>

			<ul className="nav nav-tabs" role="tablist">
				<li className="nav-item">
					<a className="nav-link active" aria-current="page" data-bs-toggle="tab"
					   data-bs-target="#attrs-effective">Effective</a>
				</li>
				<li className="nav-item">
					<a className="nav-link" data-bs-toggle="tab" data-bs-target="#attrs-default">Default</a>
				</li>
				<li className="nav-item">
					<a className="nav-link" data-bs-toggle="tab" data-bs-target="#attrs-normal">Normal</a>
				</li>
				<li className="nav-item">
					<a className="nav-link" data-bs-toggle="tab" data-bs-target="#attrs-override">Override</a>
				</li>
				<li className="nav-item">
					<a className="nav-link" data-bs-toggle="tab" data-bs-target="#attrs-automatic">Automatic</a>
				</li>
				<li className="nav-item">
					<a className="nav-link" data-bs-toggle="tab" data-bs-target="#attrs-full">Full</a>
				</li>
			</ul>

			<div className="tab-content">
				<div className="tab-pane" id="attrs-default">
					<AttrTable level={"default"} attributes={data.default}/>
				</div>
				<div className="tab-pane" id="attrs-override">
					<AttrTable level={"default"} attributes={data.default}/>
				</div>
			</div>

		</>
	)
}
