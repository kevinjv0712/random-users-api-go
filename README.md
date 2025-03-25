# Random Users API (Golang)

Esta API REST escrita en Go consulta la API pública [randomuser.me](https://randomuser.me) para obtener un conjunto de **15,000 usuarios únicos**, procesarlos concurrentemente y entregarlos como respuesta estructurada en JSON.

---

##  Características

-  **Concurrencia**: Utiliza goroutines para hacer múltiples solicitudes al mismo tiempo.
-  **Alto rendimiento**: La respuesta debe generarse en menos de 2.25 segundos.
-  **Caché local**: Guarda los datos en un archivo JSON para evitar repetir llamadas innecesarias a la API.
-  **Estadísticas incluidas**: Retorna el número de usuarios hombres y mujeres.

---

##  Requisitos de la API

Cada usuario contiene:

- `gender`: Género
- `first_name`: Primer nombre
- `last_name`: Primer apellido
- `email`: Correo electrónico
- `city`: Ciudad
- `country`: País
- `uuid`: Identificador único

Además, la respuesta incluye:

- `male_count`: Total de hombres
- `female_count`: Total de mujeres
- `total_users`: Total de usuarios
- `execution_time`: Tiempo que tomó generar la respuesta

---

##  ¿Por qué se implementa caché?

Consultar 15,000 usuarios en cada petición es costoso en tiempo y recursos.  
La caché permite:

- Acelerar las respuestas (lectura desde disco es más rápida).
- Evitar sobrecargar la API externa.
- Reducir el tiempo de espera para el cliente.

---

##  Cómo ejecutar el proyecto

```bash
go run main.go

Accede en tu navegador o con Postman a:
http://localhost:8080/users
