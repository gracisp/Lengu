# Control 2: Lenguajes de Programación

**Profesor:** Alonso Inostrosa Psijas  
**Fecha de Entrega:** 28/10/2025  
**Integrantes:**
- *Romily Barria*
- *Camilo Chavol*
- *Graciela Suárez*

## Repositorio
El código fuente (archivo 'tarea.go') se encuentra respaldado en el siguiente repositorio:  
**https://github.com/CamiloChavol/LenguajesDeProgramaci-n.git**


## Instrucciones de Ejecución

### Prerrequisitos 
Asegurarse de tener la version de GO 1.16 o superior instalado en el sistema que se utilizará.

### Ejecución

En la terminal, si ya se encuentra en la carpeta *Tarea2LP*, el programa se ejecuta a través del siguiente comando:    
``` 
go run tarea.go [n] [umbral] [dificultadPoW] [maxPrimos] [nombre_archivo]
```
Donde:
- `n`: Representa la dimensión de las dos matrices.  
- `umbral`: Valor utilizado para determinar si se ejecuta una u otra rama.  
- `dificultadPoW`: Representa la dificultad del Proof of Work.  
- `maxPrimos`: Límite de búsqueda de números primos.   
- `nombre_archivo`: Nombre del archivo de sálida donde se guarda el reporte final junto a la extensión de este.

En caso de que se este ejecutando en una terminal que no se encuentra directamente en la carpeta en la que se encuentra el archivo, se debe ejecutar el siguiente comando:  
``` 
cd [ruta de acceso]
```
Donde:
- `ruta de acceso`: Corresponde a la ruta de acceso en la que se encuentra la carpeta Tarea2LP*   

Luego de esto se debe ingresar el comando que ya se explico brevemente:  
``` 
go run tarea.go [n] [umbral] [dificultadPoW] [maxPrimos] [nombre_archivo]
```


## Descripción 

El desarrollo se realizó en **lenguaje GO**, utilizando **Gorountines** y **Canales (Channels)** para la comunicación y sincronización, implementando el patrón de **ejecuciión especulativa**, esto para ejecutar tareas en paralelo mientras se calcula una condición lógica costosa.

- **Rama A:** Ejecuta `SimularProofOfWork(data, dificultad)`  
- **Rama B:** Ejecuta `EncontrarPrimos(max)`  
- **Condición:** Determinada por `CalcularTrazaDeProductoDeMatrices(n)`

Mientras la traza se calcula, ambas ramas se ejecutan de manera especulativa.  

Cuando se conoce el resultado de la condición, se selecciona la rama ganadora y se **cancela la perdedora** mediante `context.Context`, optimizando recursos.  

La carpeta **"Tarea2LP"** que contiene este archivo llamado *"readme.md"*, también contiene los archivos *"tarea.go"* que provee el código del programa, y el archivo *"reporte.txt"* que contiene el análisis de rendimiento.

## Lógica del Programa

1. Se ejecutan en paralelo las ramas **A**, **B** y la **decisión**.  
2. Una vez determinada la traza, el programa:
   - Mantiene la rama ganadora.  
   - Cancela la rama perdedora.  
3. Se comparan los tiempos con una **ejecución secuencial** equivalente.  
4. Se calcula el **Speedup** para medir la mejora.  
5. Los resultados se guardan en el archivo de salida (`reporte.txt`).

## Análisis de Rendimiento
Para realizar el análisis de rendimiento se realizó la **Ejecución especulativa** dos veces en distintos escenarios, los parametros y resultados que se obtuvieron fueron los siguientes: 

### Escenario 1
---
- `n`: 30
- `umbral`: 5000
- `dificultadPoW`: 5
- `maxPrimos`: 500000  
- `nombre_archivo`: Escenario1.txt

| Estrategia     | Tiempo Promedio (ms) |
|----------------|---------------------|
| Secuencial     | 709.457386          |
| Especulativo   | 832.35559           |

**Speedup obtenido:** `0.85`  
> En esta prueba, la ejecución especulativa fue ligeramente más lenta que la secuencial, con una pérdida de rendimiento del 15%. Esto se debe al costo adicional de lanzar y sincronizar las Goroutines, que en este caso superó cualquier beneficio del paralelismo.


### Escenario 2
---
- `n`: 30
- `umbral`: 7500
- `dificultadPoW`: 5
- `maxPrimos`: 500000  
- `nombre_archivo`: Escenario2.txt   
   

| Estrategia     | Tiempo Promedio (ms) |
|----------------|---------------------|
| Secuencial     | 767.24631           |
| Especulativo   | 861.690883          |

**Speedup obtenido:** `0.89`  
> En esta prueba, la ejecución especulativa mostró una mejora relativa comparada con el escenario 1, pero aún así fue 11% más lenta que la secuencial. El cambio de umbral redujo ligeramente la penalización por concurrencia, pero no fue suficiente para alcanzar un speedup positivo.

## Gráficas con los resultados obtenidos
Las gráficas muestran la comparación del rendimiento realizada en los escenarios 1 y 2, cuyos parametros ya fueron definidos previamente, cabe mencionar que todos los parametros se mantuvieron fijos a excepción del umbral.

![Gráfica de Análisis Secuencial y Especulativo](img/Tiempo%20Promedio%20Secuencial%20Vs%20Especulativo.png
)
**Comportamiento en ambos escenarios**: En ambos escenarios, el tiempo promedio de ejecución especulativa es mayor que el tiempo de ejecución Secuencial.
- **Escenario 1 (umbral = 5000)**: La diferencia es notable (709 ms Vs 832 ms), lo que sugiere que el costo de gestión de la concurrencia (lanzar Gorountines y usar `context.Context` para cancelación) domina la ejecución.
- **Escenario 2 (umbral = 7500)**: La direferencia de tiempo se reduce (767 ms Vs 861 s). Este incremento en el tiempo sugiere que, dado que el tiempo de decisión (`CalcularTrazaDeProductoDeMatrices` con n = 30) es muy bajo, la ejecución especulativa presenta un costo de gestión de la concurrencia que no es compensado por el ahorro de tiempo, resultando en un tiempo total de ejecución más largo.

![Gráfica de Speedup](img/SpeedUp.png)
**Speedup < 1**: La gráfica de Speedup confirma que en niguno de los escenarios se logró una mejora de Velocidad (Speedus > 1).
- **Escenario 1 (umbral = 5000)**: El Speedup de 0.85 es el valor mas bajo, indicando que el costo de gestión es más pronunciado en este caso, ya que hay una pérdida de rendimiento del 15%.

- **Escenario 2 (umbral = 7500)**: El Speedup se sitúa es 0.89. Aunque este valor es ligeramente superior al 0.85 del escenario 1 y muestra una ligera mejora pasando de 4%, confirma que el costo de gestión domina la ejecución en ambas pruebas, ya que sigue siendo inferior a 1.

**Análisis de TEndencia**: La tendencia decreciente del Speedup de (0.85 a 0.98) sugiere que incrementar la complejidad sin ajustar el balance computacional empeora la eficiencia de la especulación.

## Conclusión Gráfica
Las gráficas demuestran que la ejecución especulativa requiere un balance óptimo entre el tiempo de decisión y las ramas conputacionales. En estos escenarios, la función de decisión (`CalcularTrazaDeProductoDeMatrices` con n = 30) fue demasiado rápida comprada con las tareas especulativas, haciendo que el costo adicional de la concurrencia (Costo de gestión) superara cualquier potencial beneficio.


## Conclusiones

- Se implementó correctamente el **patrón de ejecución especulativa** usando `goroutines`, `channels` y `context.Context` para control de cancelación.  
- Se evidenció cómo el rendimiento depende del tiempo relativo entre las ramas y la función de decisión.  
- Aunque en este experimento el **Speedup fue < 1**, el modelo demuestra el potencial de la ejecución especulativa en casos donde la condición sea significativamente costosa.  
- Se cumplieron los requerimientos de entrada, sincronización, y análisis de rendimiento solicitados en el control.
