Helpful doc:
 [Expose Pod information to containers through Env variables](https://kubernetes.io/docs/tasks/inject-data-application/environment-variable-expose-pod-information/)
[Downward API](https://kubernetes.io/docs/concepts/workloads/pods/downward-api/#:~:text=There%20are%20two%20ways%20to,are%20called%20the%20downward%20API.)

```go
func (l kuberaes) deleteDownloadBackupPodResources(namespace string, nameLabelCommon string) {
	ctx := context.TODO()
	l.clientset.AppsV1().Deployments(namespace).Delete(ctx, nameLabelCommon, metav1.DeleteOptions{})
	l.clientset.CoreV1().Services(namespace).Delete(ctx, nameLabelCommon, metav1.DeleteOptions{})
}

func (l kuberaes) downloadDbBackup(inp DbBackupDownloadReq, w http.ResponseWriter) {
	ctx := context.TODO()
	pvcName := inp.Name + "-db-cronjob-backup"
	nameLabelCommon := "download-" + inp.BackupName + "-backup"
	namespace := inp.Namespace + "-ns"

	log.Println(pvcName)
	log.Println(nameLabelCommon)
	log.Println(namespace)
	log.Println(inp.BackupName)

	job, jobErr := l.clientset.CoreV1().Pods(namespace).Get(ctx, inp.BackupName, metav1.GetOptions{})

	if jobErr != nil {
		resBody := ResponseEntityTemplate{
			Message: "Backup not found",
			Status:  http.StatusNotFound,
		}
		json.NewEncoder(w).Encode(resBody)
		return
	}

	backupId := job.ObjectMeta.Name

	log.Println("BackupId: " + backupId)

	_, err := l.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		resBody := ResponseEntityTemplate{
			Message: "DB dump pvc not found",
			Status:  http.StatusNotFound,
		}
		json.NewEncoder(w).Encode(resBody)
		return
	}

	labels := map[string]string{
		"backup": nameLabelCommon,
	}

	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameLabelCommon,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  nameLabelCommon,
							Image: "aasourav/db-dump-download-helper",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: int32(8035),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      pvcName,
									MountPath: "/tmp", // Inside container location.
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: pvcName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}

	dep, err := l.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})

	if err != nil {
		resBody := ResponseEntityTemplate{
			Message: "Failed to run API server",
			Status:  http.StatusInternalServerError,
		}
		json.NewEncoder(w).Encode(resBody)
		return
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: nameLabelCommon,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       8035,
					TargetPort: intstr.FromInt32(8035),
				},
			},
		},
	}

	_, err = l.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})

	if err != nil {
		resBody := ResponseEntityTemplate{
			Message: "Failed to run API server",
			Status:  http.StatusInternalServerError,
		}
		json.NewEncoder(w).Encode(resBody)
		return
	}

	// wait for preparing api server
	time.Sleep(5 * time.Second)
	timeout := time.Now()
	for {
		if dep.Status.AvailableReplicas == dep.Status.ReadyReplicas {
			time.Sleep(2 * time.Second)
			break
		}
		if time.Since(timeout) > 60*time.Second {
			resBody := ResponseEntityTemplate{
				Message: "Request timeout",
				Status:  http.StatusNotFound,
			}
			json.NewEncoder(w).Encode(resBody)
			l.deleteDownloadBackupPodResources(namespace, nameLabelCommon)
			return
		}
		time.Sleep(2 * time.Second)
	}

	fileURL := fmt.Sprintf("http://%v.%v.svc.cluster.local:8035/download/%v.tar.gz", nameLabelCommon, namespace, backupId)

	// Make a request to the file URL
	resp, err := http.Get(fileURL)
	if err != nil {
		resBody := ResponseEntityTemplate{
			Message: err.Error(),
			Status:  http.StatusNotFound,
		}
		json.NewEncoder(w).Encode(resBody)
		l.deleteDownloadBackupPodResources(namespace, nameLabelCommon)
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		resBody := ResponseEntityTemplate{
			Message: "File not found",
			Status:  http.StatusNotFound,
		}
		json.NewEncoder(w).Encode(resBody)
		l.deleteDownloadBackupPodResources(namespace, nameLabelCommon)
		return
	}

	// Set headers
	w.Header().Set("Content-Disposition", "attachment; filename="+inp.BackupName+".tar.gz")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))

	// Stream the content to the response writer
	_, err = io.Copy(w, resp.Body)

	if err != nil {
		if resp.StatusCode != http.StatusOK {
			resBody := ResponseEntityTemplate{
				Message: "Internal server error",
				Status:  http.StatusInternalServerError,
			}
			json.NewEncoder(w).Encode(resBody)
		}
	}

	defer func() {
		l.deleteDownloadBackupPodResources(namespace, nameLabelCommon)
	}()
}
```