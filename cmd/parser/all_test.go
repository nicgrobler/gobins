package main

import (
	"encoding/json"
	"testing"
	"text/template"
)

func findObjectIndex(name string, files []string) (int, bool) {
	// returns the index of name
	for i, f := range files {
		if f == name {
			return i, true
		}
	}
	return 0, false
}

func generateTemplate() *template.Template {
	s := `
		{{ $data := . }}
	
		{{ $modifiedProjectName := replace $data.ProjectName "-" "_" }}
    	{{ $upperCaseLineEnd := upper $modifiedProjectName }}
    	{{ $upperCaseEnv := upper $data.Environment }}
	
		[{
       
			"adgroup":"RES-{{$upperCaseEnv}}-OPSH-DEVELOPER-{{$upperCaseLineEnd}}",
			"role":"EDIT",
			"bindingname":"adgroup-edit-binding",
			"filename":"10-edit-group-rolebinding.json"
			
		},
		{
			"adgroup":"RES-{{$upperCaseEnv}}-OPSH-VIEWER-{{$upperCaseLineEnd}}",
			"role":"VIEW",
			"bindingname":"adgroup-view-binding",
			"filename":"10-view-group-rolebinding.json"
			
		},
		{
		   
			"adgroup":"RES-{{$upperCaseEnv}}-OPSH-DEPLOY-RELMAN",
			"role":"ADMIN",
			"bindingname":"adgroup-deploy-binding",
			"filename":"10-jenkins-rolebinding.json"
		   
		}]
	`
	funcMap := template.FuncMap{
		"replace": replace,
		"upper":   upper,
		"lower":   lower,
	}
	t := getTemplateFromString("raw_stream", s, funcMap)
	return t
}

func TestCheckInputValid(t *testing.T) {
	data := []byte(`{
		"projectname": "nic-test-backbase-reference",
		"role": "developer",
		"environment": "dev",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d := expectedInput{}
	err := json.Unmarshal(data, &d)
	if err != nil {
		t.Errorf("wanted %s, but got %s: \n", "nil", err.Error())
	}

	if d.Environment != "dev" {
		t.Errorf("wanted %s, but got %s: \n", "dev", d.Environment)
	}

	if d.Optionals == nil {
		t.Errorf("wanted %s, but got %s: \n", "optionals", "nil")
	}

	// complain about spaces
	badData := []byte(`{
		"projectname": "nic-test backbase-reference",
		"environment": "dev",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	want := "data contains illegal spaces"
	got := err.Error()
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	// complain about underscores
	badData = []byte(`{
		"projectname": "nic_test-backbase-reference",
		"environment": "dev",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	want = "data contains illegal underscores"
	got = err.Error()
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	// should autoformat the data
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err != nil {
		t.Errorf("wanted %s, but got %s: \n", "nil", err.Error())
	}

	if d.ProjectName != "nic-test-backbase-reference" {
		t.Errorf("wanted %v, but got %v: \n", "nic-test-backbase-reference", d.ProjectName)
	}

	if d.Environment != "dev" {
		t.Errorf("wanted %v, but got %v: \n", "dev", d.Environment)
	}

	// should complain about invalid name in optionals
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memooory",
						"count":1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "optional name entry is invalid: memooory" {
		t.Errorf("wanted %s, but got %s: \n", "an error", err.Error())
	}

	// should complain about invalid unit in optionals
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1,
						"unit":"Giz"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "optional unit entry is invalid: Giz" {
		t.Errorf("wanted %s, but got %s: \n", "optional unit entry is invalid: Giz", err.Error())
	}

	// should complain about missing unit in optionals
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"storage",
						"count":1
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "invalid or missing unit for: storage" {
		t.Errorf("wanted %s, but got %s: \n", "invalid or missing unit for: storage", err.Error())
	}

	// should complain about invalid count in optionals with type error
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":1.1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "json: cannot unmarshal number 1.1 into Go struct field optionalObject.Optionals.count of type int" {
		t.Errorf("wanted %s, but got %s: \n", "json: cannot unmarshal number 1.1 into Go struct field optionalObject.Optionals.count of type int", err.Error())
	}

	// should complain about invalid count in optionals with type error
	badData = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1
					},
					{
						"name":"memory",
						"count":"1",
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(badData, &d)
	if err == nil {
		t.Errorf("wanted %s, but got %s: \n", "an error", "nil")
	}
	if err.Error() != "json: cannot unmarshal string into Go struct field optionalObject.Optionals.count of type int" {
		t.Errorf("wanted %s, but got %s: \n", "json: cannot unmarshal string into Go struct field optionalObject.Optionals.count of type int", err.Error())
	}

	// should complain about invalid count in optionals with type error
	data = []byte(`{
		"projectname": "NIC-test-backbase-reference",
		"environment": "DEV",
		"optionals":[
					{
						"name":"cpu",
						"count": 1000,
						"unit": "m"
					},
					{
						"name":"memory",
						"count":1,
						"unit":"Gi"
					},
					{
						"name":"volumes",
						"count":2
					}
		]
	}`)
	d = expectedInput{}
	err = json.Unmarshal(data, &d)
	if err != nil {
		t.Errorf("wanted %s, but got %s: \n", "nil", err.Error())
	}
	if d.Optionals[0].Count.int != 1000 || d.Optionals[0].Unit.string != "m" {
		t.Errorf("wanted %s, but got %s: \n", "should be equal", "are not equal")
	}

}

func TestValidUnit(t *testing.T) {

	want := false
	got := validUnit("gb")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validUnit("Mi")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}
}

func TestValidName(t *testing.T) {
	want := false
	got := validName("cpus")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validName("cpu")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = false
	got = validName("Memory")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validName("volumes")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = true
	got = validName("storage")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}

	want = false
	got = validName("disk")
	if got != want {
		t.Errorf("wanted %v, but got %v: \n", want, got)
	}
}

func TestCreateNewProjectObject(t *testing.T) {

	expectedBytes := []byte(`[{"content":{"apiVersion":"project.openshift.io/v1","kind":"Project","metadata":{"name":"boogie-test"}},"filename":"1-project.json"]`)

	i := expectedInput{ProjectName: "boogie-test"}
	c := config{
		flatOutput:          true,
		usefileContentInput: true,
		fileContent: `{{ $data := . }}

		{{ $lowerProjectName := lower $data.ProjectName }}
		
		[{
			"filename": "1-project.json",
			"content": {
			  "kind": "Project",
			  "apiVersion": "project.openshift.io/v1",
			  "metadata": {
				"name": "{{$lowerProjectName}}"
			  }
			}
		}]`,
	}
	gotBytes, err := c.process(&i)

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
}

func TestCreateNewNetworkPolicyObject(t *testing.T) {

	expectedBytes := []byte(`[{"content":{"apiVersion":"networking.k8s.io/v1","kind":"NetworkPolicy","metadata":{"name":"default-deny-all","namespace":"boogie-test"},"spec":{"podSelector":{},"policyTypes":["Ingress"]}},"filename":"10-networkpolicy.json"]`)

	i := expectedInput{ProjectName: "boogie-test"}
	c := config{
		flatOutput:          true,
		usefileContentInput: true,
		fileContent: `{{ $data := . }}

		{{ $lowerCaseProjectName := lower $data.ProjectName }}
		
		[{
		  
		"filename":"10-networkpolicy.json",
		"content":
		  {
		  "apiVersion": "networking.k8s.io/v1",
			"kind": "NetworkPolicy",
			"metadata": {
			  "name": "default-deny-all",
			  "namespace": "{{$lowerCaseProjectName}}"
			},
			"spec": {
			  "podSelector": {},
			  "policyTypes": [
				"Ingress"
			  ]
			}
		 }
		}]`,
	}
	gotBytes, err := c.process(&i)

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

}

func TestCreateNewEgressNetworkPolicyObject(t *testing.T) {

	expectedBytes := []byte(`[{"content":{"apiVersion":"network.openshift.io/v1","kind":"EgressNetworkPolicy","metadata":{"name":"default-egress","namespace":"boogie-test"},"spec":{"egress":[{"to":{"cidrSelector":"0.0.0.0/0"},"type":"Deny"}]}},"filename":"10-egress-networkpolicy.json"]`)

	i := expectedInput{ProjectName: "boogie-test"}
	c := config{
		flatOutput:          true,
		usefileContentInput: true,
		fileContent: `{{ $data := . }}

		{{ $lowerCaseProjectName := lower $data.ProjectName }}
		
		[{
			"filename": "10-egress-networkpolicy.json",
			"content": {
			  "kind": "EgressNetworkPolicy",
			  "apiVersion": "network.openshift.io/v1",
			  "metadata": {
				"name": "default-egress",
				"namespace": "{{$lowerCaseProjectName}}"
			  },
			  "spec": {
				"egress": [
				  {
					"type": "Deny",
					"to": {
					  "cidrSelector": "0.0.0.0/0"
					}
				  }
				]
			  }
			} 
		}]`,
	}
	gotBytes, err := c.process(&i)

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

}

func TestCreateNewRoleBindingObject(t *testing.T) {

	expectedBytes := []byte(`[{"content":{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"RoleBinding","metadata":{"name":"adgroup-edit-binding","namespace":"boogie-test"},"roleRef":{"apiGroup":"rbac.authorization.k8s.io","kind":"ClusterRole","name":"edit"},"subjects":[{"apiGroup":"rbac.authorization.k8s.io","kind":"Group","name":"RES--OPSH-DEVELOPER-BOOGIE_TEST"}]},"filename":"10-edit-group-rolebinding.json"},{"content":{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"RoleBinding","metadata":{"name":"adgroup-view-binding","namespace":"boogie-test"},"roleRef":{"apiGroup":"rbac.authorization.k8s.io","kind":"ClusterRole","name":"view"},"subjects":[{"apiGroup":"rbac.authorization.k8s.io","kind":"Group","name":"RES--OPSH-VIEWER-BOOGIE_TEST"}]},"filename":"10-view-group-rolebinding.json"},{"content":{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"RoleBinding","metadata":{"name":"adgroup-deploy-binding","namespace":"boogie-test"},"roleRef":{"apiGroup":"rbac.authorization.k8s.io","kind":"ClusterRole","name":"admin"},"subjects":[{"apiGroup":"rbac.authorization.k8s.io","kind":"Group","name":"RES--OPSH-DEPLOY-RELMAN"}]},"filename":"10-jenkins-rolebinding.json"},{"content":{"apiVersion":"rbac.authorization.k8s.io/v1","kind":"RoleBinding","metadata":{"name":"adgroup-manage-binding","namespace":"boogie-test"},"roleRef":{"apiGroup":"rbac.authorization.k8s.io","kind":"ClusterRole","name":"deploy"},"subjects":[{"apiGroup":"rbac.authorization.k8s.io","kind":"Group","name":"RES--OPSH-MANAGE-RELMAN"}]},"filename":"10-default-rolebinding.json"]`)

	i := expectedInput{ProjectName: "boogie-test"}
	c := config{
		flatOutput:          true,
		usefileContentInput: true,
		fileContent: `    {{ $data := . }}

		{{ $modifiedProjectName := replace $data.ProjectName "-" "_" }}
		{{ $upperCaseLineEnd := upper $modifiedProjectName }}
		{{ $upperCaseEnv := upper $data.Environment }}
		{{ $lowerProjectName := lower $data.ProjectName }}
	
	[{
		"filename": "10-edit-group-rolebinding.json",
		"content": {
		  "kind": "RoleBinding",
		  "apiVersion": "rbac.authorization.k8s.io/v1",
		  "metadata": {
			"name": "adgroup-edit-binding",
			"namespace": "{{$lowerProjectName}}"
		  },
		  "subjects": [
			{
			  "kind": "Group",
			  "apiGroup": "rbac.authorization.k8s.io",
			  "name": "RES-{{$upperCaseEnv}}-OPSH-DEVELOPER-{{$upperCaseLineEnd}}"
			}
		  ],
		  "roleRef": {
			"kind": "ClusterRole",
			"apiGroup": "rbac.authorization.k8s.io",
			"name": "edit"
		  }
		}
	  },
	  {
		"filename": "10-view-group-rolebinding.json",
		"content": {
		  "kind": "RoleBinding",
		  "apiVersion": "rbac.authorization.k8s.io/v1",
		  "metadata": {
			"name": "adgroup-view-binding",
			"namespace": "{{$lowerProjectName}}"
		  },
		  "subjects": [
			{
			  "kind": "Group",
			  "apiGroup": "rbac.authorization.k8s.io",
			  "name": "RES-{{$upperCaseEnv}}-OPSH-VIEWER-{{$upperCaseLineEnd}}"
			}
		  ],
		  "roleRef": {
			"kind": "ClusterRole",
			"apiGroup": "rbac.authorization.k8s.io",
			"name": "view"
		  }
		}
	  },
	  {
		"filename": "10-jenkins-rolebinding.json",
		"content": {
		  "kind": "RoleBinding",
		  "apiVersion": "rbac.authorization.k8s.io/v1",
		  "metadata": {
			"name": "adgroup-deploy-binding",
			"namespace": "{{$lowerProjectName}}"
		  },
		  "subjects": [
			{
			  "kind": "Group",
			  "apiGroup": "rbac.authorization.k8s.io",
			  "name": "RES-{{$upperCaseEnv}}-OPSH-DEPLOY-RELMAN"
			}
		  ],
		  "roleRef": {
			"kind": "ClusterRole",
			"apiGroup": "rbac.authorization.k8s.io",
			"name": "admin"
		  }
		}
	  },
	  {
		"filename": "10-default-rolebinding.json",
		"content": {
		  "kind": "RoleBinding",
		  "apiVersion": "rbac.authorization.k8s.io/v1",
		  "metadata": {
			"name": "adgroup-manage-binding",
			"namespace": "{{$lowerProjectName}}"
		  },
		  "subjects": [
			{
			  "kind": "Group",
			  "apiGroup": "rbac.authorization.k8s.io",
			  "name": "RES-{{$upperCaseEnv}}-OPSH-MANAGE-RELMAN"
			}
		  ],
		  "roleRef": {
			"kind": "ClusterRole",
			"apiGroup": "rbac.authorization.k8s.io",
			"name": "deploy"
		  }
		}
	}]`,
	}
	gotBytes, err := c.process(&i)

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

}

func TestCreateNewLimitsObject(t *testing.T) {
	expectedBytes := []byte(`[{"content":{"apiVersion":"v1","kind":"ResourceQuota","metadata":{"name":"default-quotas","namespace":"boogie-test"},"spec":{"hard":{"limits.cpu":2,"limits.memory":"1Gi","persistentvolumeclaims":3,"requests.storage":"100Gi"}}},"filename":"10-quotas.json"]`)

	o := []optionalObject{
		optionalObject{
			Name:  oName{"cpu"},
			Count: oCount{2},
		},
		optionalObject{
			Name:  oName{"memory"},
			Count: oCount{1},
			Unit:  oUnit{"Gi"},
		},
		optionalObject{
			Name:  oName{"volumes"},
			Count: oCount{3},
		},
		optionalObject{
			Name:  oName{"storage"},
			Count: oCount{100},
			Unit:  oUnit{"Gi"},
		},
	}

	i := expectedInput{ProjectName: "boogie-test", Optionals: o}
	c := config{
		flatOutput:          true,
		usefileContentInput: true,
		fileContent: `{{ $data := . }}

		{{ $lowerProjectName := lower $data.ProjectName }}
		
		{{ $cpu := getCPU $data "100m" }}
		{{ $mem := getMEM $data "100Mi" }}
		{{ $pvc := getPVC $data 1 }}
		{{ $storage := getStorage $data "1Gi" }}
		
		
		[
		  {
			"filename": "10-quotas.json",
		   "content": {
			  "kind": "ResourceQuota",
			  "apiVersion": "v1",
			  "metadata": {
				"name": "default-quotas",
				"namespace": "{{$lowerProjectName}}"
			  },
			  "spec": {
				"hard": {
				  "limits.cpu": {{$cpu}},
				  "limits.memory": {{$mem}},
				  "persistentvolumeclaims": {{$pvc}},
				  "requests.storage": {{$storage}}
				}
			  }
			}
		
		  }]`,
	}
	gotBytes, err := c.process(&i)

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

	expectedBytes = []byte(`[{"content":{"apiVersion":"v1","kind":"ResourceQuota","metadata":{"name":"default-quotas","namespace":"boogie-test"},"spec":{"hard":{"limits.cpu":1,"limits.memory":"5Gi","persistentvolumeclaims":1,"requests.storage":"5Gi"}}},"filename":"10-quotas.json"]`)

	o = []optionalObject{
		optionalObject{
			Name:  oName{"cpu"},
			Count: oCount{1},
		},
		optionalObject{
			Name:  oName{"memory"},
			Count: oCount{5},
			Unit:  oUnit{"Gi"},
		},
		optionalObject{
			Name:  oName{"storage"},
			Count: oCount{5},
			Unit:  oUnit{"Gi"},
		}}

	i = expectedInput{ProjectName: "boogie-test", Environment: "dev", Optionals: o}

	gotBytes, err = c.process(&i)

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}
	/*
		expectedBytes = []byte(`{"kind":"","apiVersion":"","metadata":{"name":""},"spec":{"hard":{}}}`)

		i = expectedInput{ProjectName: "boogie-test", Environment: "dev"}

		gotBytes, err = c.process(&i)

		if err != nil {
			t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
		}

		if string(expectedBytes) != string(gotBytes) {
			t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
		}
	*/
}

func TestCreateNewLimitsObjectCPU(t *testing.T) {
	expectedBytes := []byte(`[{"content":{"apiVersion":"v1","kind":"ResourceQuota","metadata":{"name":"default-quotas","namespace":"boogie-test"},"spec":{"hard":{"limits.cpu":"200m","limits.memory":"1Gi","persistentvolumeclaims":3,"requests.storage":"1Gi"}}},"filename":"10-quotas.json"]`)

	o := []optionalObject{
		optionalObject{
			Name:  oName{"cpu"},
			Count: oCount{200},
			Unit:  oUnit{"m"},
		},
		optionalObject{
			Name:  oName{"memory"},
			Count: oCount{1},
			Unit:  oUnit{"Gi"},
		},
		optionalObject{
			Name:  oName{"volumes"},
			Count: oCount{3},
		}}

	i := expectedInput{ProjectName: "boogie-test", Environment: "dev", Optionals: o}

	c := config{
		flatOutput:          true,
		usefileContentInput: true,
		fileContent: `{{ $data := . }}

		{{ $lowerProjectName := lower $data.ProjectName }}
		
		{{ $cpu := getCPU $data "100m" }}
		{{ $mem := getMEM $data "100Mi" }}
		{{ $pvc := getPVC $data 1 }}
		{{ $storage := getStorage $data "1Gi" }}
		
		
		[
		  {
			"filename": "10-quotas.json",
		   "content": {
			  "kind": "ResourceQuota",
			  "apiVersion": "v1",
			  "metadata": {
				"name": "default-quotas",
				"namespace": "{{$lowerProjectName}}"
			  },
			  "spec": {
				"hard": {
				  "limits.cpu": {{$cpu}},
				  "limits.memory": {{$mem}},
				  "persistentvolumeclaims": {{$pvc}},
				  "requests.storage": {{$storage}}
				}
			  }
			}
		
		  }
		]`,
	}
	gotBytes, err := c.process(&i)

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

}

func TestShowDefaultQuotas(t *testing.T) {
	expectedBytes := []byte(`[{"content":{"apiVersion":"v1","kind":"ResourceQuota","metadata":{"name":"default-quotas","namespace":"show-only"},"spec":{"hard":{"limits.cpu":"100m","limits.memory":"100Mi","persistentvolumeclaims":1,"requests.storage":"1Gi"}}},"filename":"10-quotas.json"]`)

	c := config{
		flatOutput:          true,
		usefileContentInput: true,
		fileContent: `{{ $data := . }}

		{{ $lowerProjectName := lower $data.ProjectName }}
		
		{{ $cpu := getCPU $data "100m" }}
		{{ $mem := getMEM $data "100Mi" }}
		{{ $pvc := getPVC $data 1 }}
		{{ $storage := getStorage $data "1Gi" }}
		
		
		[
		  {
			"filename": "10-quotas.json",
		   "content": {
			  "kind": "ResourceQuota",
			  "apiVersion": "v1",
			  "metadata": {
				"name": "default-quotas",
				"namespace": "{{$lowerProjectName}}"
			  },
			  "spec": {
				"hard": {
				  "limits.cpu": {{$cpu}},
				  "limits.memory": {{$mem}},
				  "persistentvolumeclaims": {{$pvc}},
				  "requests.storage": {{$storage}}
				}
			  }
			}
		
		  }
		]`,
	}
	gotBytes, err := c.show()

	if err != nil {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", "no error", err.Error())
	}

	if string(expectedBytes) != string(gotBytes) {
		t.Errorf("wanted \n%s, \nbut got \n%s \n", expectedBytes, gotBytes)
	}

}
