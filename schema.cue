// this is a multiline comment
// that introduces the project.
//

// this is a new value
id: *"transactions" | string

// sets the replicas for the service
replicas!: *2 | int

// replicas max
max: replicas * 3

// new value
owner!: string & =~".+@sap.com"

// counter value
counter!: int

// should tls be enabled for the service 
enableTLS: *true | bool

// this is a required field with of type string with a constraint
redis_url!: string & =~"^https://.+:6379$"

// gate is an endpoint that will be checked at the given interval 
gate!: {
	// url 
	url: "test"

	// interval in miliseconds
	interval: *1000 | int

	// data
	data: {
		// forename
		forname: *"piaras" | string
		// surname
		surname: string

		// address
		address: {
			// home place
			home: string

			// phone number
			phone: string
		}
	}

}

// this is an optional field
labels?: {[string]: string}

// annotations can be added optionally
annotations?: {[string]: {}}

annotations: {
	[string]: {
		age: int
	}
}

// count is a series of numbers
// that have very little relevance to anyone
count: [1, 2, 3, 4, 5]
