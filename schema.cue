// this is a multiline comment
// that describes the configuration.

// this is the schema version
#SchemaVersion: "v1.0.0"

// this is a new value
service_id: *"transactions-svc" | string 

// sets the replicas for the service
replicas!: *2 | int 

// replicas max
max_replicas: replicas * 5 @private(true)

// this is a managed field that can be edited. the lifecycle of this field is 
// managed by the system and if the schema_version changes
// then the system will manage the update of this field. 
//
// should tls be enabled for the service 
enable_tls: *true | bool @manage(id=system)

// this is a required field with of type string with a constraint
redis_url: string & =~"^https://.+:6379$"
