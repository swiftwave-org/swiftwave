package dockermanager

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
)

// Service manager

// Create a new service
func (m Manager) CreateService(name string, replicas uint64) {

	 d, err := m.client.ServiceCreate(m.ctx, swarm.ServiceSpec{
		// Set name of the service
		Annotations: swarm.Annotations{
			Name: name,
		},
		// Set task template
		TaskTemplate: swarm.TaskSpec{
			// Set container spec
			ContainerSpec: &swarm.ContainerSpec{
				Image: "tanmoysrt/minc:v2",
				// Command: []string{
				// 	"/bin/sh","-c", "sleep 7200",
				// },
				Hosts: []string{},
				
				Env: []string{},
				Mounts: []mount.Mount{},
			},
			// Set network name
			Networks: []swarm.NetworkAttachmentConfig{
				swarm.NetworkAttachmentConfig{
					// TODO: update it
					Target: "swarm-network",
				},
			},
		},
		// allow replicated service
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &replicas,
			},
		},
		// constant endpoint
		EndpointSpec: &swarm.EndpointSpec{
			Mode: swarm.ResolutionModeDNSRR,
		},
	}, types.ServiceCreateOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Println(d.ID)
	
}

// Update a service
func (m Manager) UpdateService(name string, id string, replicas uint64) {

	d, err := m.client.ServiceUpdate(m.ctx, id, swarm.Version{
		Index: 6622,
	}, swarm.ServiceSpec{
	   // Set name of the service
	   Annotations: swarm.Annotations{
		   Name: name,
	   },
	   // Set task template
	   TaskTemplate: swarm.TaskSpec{
		   // Set container spec
		   ContainerSpec: &swarm.ContainerSpec{
			   Image: "tanmoysrt/minc:v2",
			   // Command: []string{
			   // 	"/bin/sh","-c", "sleep 7200",
			   // },
			   Hosts: []string{},
			   
			   Env: []string{},
			   Mounts: []mount.Mount{},
		   },
		   // Set network name
		   Networks: []swarm.NetworkAttachmentConfig{
			   swarm.NetworkAttachmentConfig{
				   // TODO: update it
				   Target: "swarm-network",
			   },
		   },
	   },
	   // allow replicated service
	   Mode: swarm.ServiceMode{
		   Replicated: &swarm.ReplicatedService{
			   Replicas: &replicas,
		   },
	   },
	   // constant endpoint
	   EndpointSpec: &swarm.EndpointSpec{
		   Mode: swarm.ResolutionModeDNSRR,
	   },
   }, types.ServiceUpdateOptions{})

   fmt.Println(d)
   fmt.Println(err)
   if err != nil {
	   panic(err)
   }
}

// Remove a service

// Get status of a service
// -- no of replicas
//    -- desired
//    -- running
//    -- failed
//    -- pending

// Get service logs

// Get service stats



