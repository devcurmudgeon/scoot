# Structures for storing the internal JobDefinition.
# We use this structure rather than scootapi's JobDefinition (in scoot.thrift)
# so that we can add data to the log and not impact the client API.

# We should use go generate to run:
# For now to Install Thrift:
#     1. Install Thrift manually `brew install thrift` ensure version is greater that 0.9.3
#     2. go get github.com/apache/thrift/lib/go/thrift
#

# To Generate files run from this (github.com/twitter/scoot/sched) directory
# (note that package_prefix must be set to the correct prefix of the included definition)
#     1. thrift -I ../bazel/execution/request/ --gen go:package_prefix=github.com/twitter/scoot/bazel/execution/request/gen-go/,thrift_import=github.com/apache/thrift/lib/go/thrift sched.thrift

include "request.thrift"

struct Command {
  1: required list<string> argv
  2: optional map<string, string> envVars
  3: optional i64 timeout
  4: required string snapshotId
}

struct TaskDefinition {
  1: required Command command
  2: optional string taskId
  3: optional request.BazelExecuteRequest bazelRequest
}

struct JobDefinition {
  1: optional string jobType,
  2: optional list<TaskDefinition> tasks,
  3: optional i32 priority
  4: optional string tag
  5: optional string basis
  6: optional string requestor
}

struct Job {
  1: required string id
  2: required JobDefinition jobDefinition
}