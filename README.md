# kubelish - mDNS service publisher for Kubernetes cluster nodes

This project implements a way to expose k8s services with external IPs as mDNS services.

The rationale behind this project is to allow k8s services to be discovered by
devices on the local network dynamically, without the need to configure static IPs
or DNS entries. This is particularly useful for IoT devices, which are often unable
to become part of a traditional service mesh but need to access services of
the k8s cluster.

In contrast to other solutions, this project exposes mDNS services (SRV records)
rather than hosts (A records) which has the advantage that the daemon can run
on multiple nodes of the same cluster that have IP addresses in the same subnet,
and offer the service on each of them. Clients will hence pick one of the nodes at
random.

## How it works

The `kubelish` daemon runs on each node of the cluster and listens for new `LoadBalancer`
services with external IPs. When a service is created or updated, the daemon checks
the annotations of the service to see if it should be exposed.

The following keys are required:

```yaml
metadata:
  annotations:
    kubelish/service-name: Example
    kubelish/service-type: _example._tcp
    kubelish/txt: Optional TXT record to be exposed along with the service on mDNS
```

If the service is annotated that way, and one of its external IP addresses is
a local IP address of the node the daemon is running on, the service will be
exposed as an mDNS service.

If a service is deleted or updated to not be exposed, the daemon will remove
the mDNS service.

## Kubernetes services with multiple ports

If a service has multiple ports, the port to be exposed as an mDNS service
has to be annotated with the same name as the mDNS service name, e.g.:

```yaml
metadata:
  annotations:
    kubelish/service-name: Example
    kubelish/service-type: _example._tcp
    kubelish/txt: Optional TXT record to be exposed along with the service on mDNS
    ...
spec:
  ...
  ports:
  - name: Example
    nodePort: 10000
    port: 9090
    protocol: TCP
    targetPort: 9090
```

## Running the daemon

The daemon must be run natively on a node of the cluster, outside of Kubernetes.

Install the daemon by using

```bash
go install github.com/holoplot/kubelish/cmd/kubelish@latest
```

Then move the binary to a suitable location in your `$PATH`, e.g. `/usr/local/bin/kubelish`.

The binary supports the following commands:

- `kubelish watch` - Runs the watcher daemon
- `kubelish add <k8s-service>` - Adds annotations to a service to expose it as an mDNS service and dumps the YAML to stdout
- `kubelish remove <k8s-service>` - Removes annotations from a service and dumps the YAML to stdout

A systemd service file is [kubelish.service](provided). Make sure you
edit the file to set the correct path to the binary and the kubeconfig file.

### Global flags

The following flags can be used with all commands:
* `--namespace` - Namespace to watch for services. Defaults to `default`.

### Environment variables

The environment variables below can be used to configure the daemon:

- `KUBECONFIG` - Path to the kubeconfig file to use. Defaults to `~/.kube/config`.

### Example

Let's assume we have a service called `my-loadbalancer` in the `default` namespace that
you want to expose as an mDNS service.

First, run the daemon on the node that has the IP address of the service:

```bash
kubelish watch
```

Then, in a second terminal, add the annotations to the service using the following command:

```bash
kubelish add my-loadbalancer --service-name example --service-type _example._tcp --txt Example \
	| kubectl apply -f -
```

The service should now be exposed as an mDNS service. You can verify this by using
the `avahi-browse` command:

```bash
avahi-browse -r _example._tcp
```

To remove the service, run the following command:

```bash
kubelish remove my-loadbalancer | kubectl apply -f -
```

## Limitations

This project is still in its early stages and has some limitations:

- Only supports Linux
- Depends on `avahi-daemon` to be installed on the host
- Only services with a single exposed port are supported
- Does not support miniKube's way of exposing services with `minikube tunnel`

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
