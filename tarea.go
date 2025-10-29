package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// =========================================================================
// PASO 1: FUNCIONES DEL ANEXO
// =========================================================================
//

// SimularProofOfWork simula la búsqueda de una prueba de trabajo de blockchain.
// Modificada para ser CANCELABLE usando un context.Context.
func SimularProofOfWork(ctx context.Context, blockData string, dificultad int) (string, int, error) {
	targetPrefix := strings.Repeat("0", dificultad)
	nonce := 0
	for {
		select {
		case <-ctx.Done(): // Si el contexto se cancela, salimos.
			return "", 0, fmt.Errorf("POW cancelado")
		default:
			// Continuamos trabajando
			data := fmt.Sprintf("%s%d", blockData, nonce)
			hashBytes := sha256.Sum256([]byte(data))
			hashString := fmt.Sprintf("%x", hashBytes)

			if strings.HasPrefix(hashString, targetPrefix) {
				return hashString, nonce, nil // Encontramos el resultado
			}
			nonce++
		}
	}
}

// EncontrarPrimos busca todos los números primos hasta un entero max.
// Modificada para ser CANCELABLE usando un context.Context.
func EncontrarPrimos(ctx context.Context, max int) ([]int, error) {
	var primes []int
	for i := 2; i < max; i++ {
		// Verificamos la cancelación en cada iteración "grande"
		select {
		case <-ctx.Done(): // Si el contexto se cancela, salimos.
			return nil, fmt.Errorf("búsqueda de primos cancelada")
		default:
			// Continuamos trabajando
			isPrime := true
			for j := 2; j*j <= i; j++ {
				if i%j == 0 {
					isPrime = false
					break
				}
			}
			if isPrime {
				primes = append(primes, i)
			}
		}
	}
	return primes, nil
}

// CalcularTrazaDeProductoDeMatrices multiplica dos matrices NXN y devuelve la traza.
// No necesita ser cancelable, ya que es la función de decisión.
func CalcularTrazaDeProductoDeMatrices(n int) int {
	// Se crean dos matrices con valores aleatorios
	m1 := make([][]int, n)
	m2 := make([][]int, n)
	for i := 0; i < n; i++ {
		m1[i] = make([]int, n)
		m2[i] = make([]int, n)
		for j := 0; j < n; j++ {
			m1[i][j] = rand.Intn(10)
			m2[i][j] = rand.Intn(10)
		}
	}

	// Se realiza la multiplicación y se calcula la traza
	trace := 0
	for i := 0; i < n; i++ {
		sum := 0
		for k := 0; k < n; k++ {
			sum += m1[i][k] * m2[k][i]
		}
		trace += sum
	}
	return trace
}

// =========================================================================
// PASO 2: ESTRUCTURAS PARA LOS RESULTADOS (Channels)
// =========================================================================

// ResultadoPoW encapsula el resultado de SimularProofOfWork
type ResultadoPoW struct {
	hash     string
	nonce    int
	duracion time.Duration
	err      error
}

// ResultadoPrimos encapsula el resultado de EncontrarPrimos
type ResultadoPrimos struct {
	primos   []int
	duracion time.Duration
	err      error
}

// ResultadoDecision encapsula el resultado de la función de decisión
type ResultadoDecision struct {
	traza    int
	duracion time.Duration
}

// =========================================================================
// PASO 3: IMPLEMENTACIÓN SECUENCIAL (Línea Base)
// =========================================================================

//
func ejecucionSecuencial(n, umbral, dificultadPoW, maxPrimos int, blockData string) (time.Duration, string) {
	inicioTotal := time.Now()

	// 1. Ejecutar la decisión
	inicioDecision := time.Now()
	traza := CalcularTrazaDeProductoDeMatrices(n)
	duracionDecision := time.Since(inicioDecision)

	var duracionComputo time.Duration
	var ramaSeleccionada string

	// 2. Tomar la decisión y SÓLO DESPUÉS ejecutar la rama ganadora
	if traza > umbral {
		ramaSeleccionada = "A (ProofOfWork)"
		inicioComputo := time.Now()
		// Usamos context.Background() porque en secuencial no hay nada que cancelar.
		_, _, _ = SimularProofOfWork(context.Background(), blockData, dificultadPoW)
		duracionComputo = time.Since(inicioComputo)
	} else {
		ramaSeleccionada = "B (EncontrarPrimos)"
		inicioComputo := time.Now()
		_, _ = EncontrarPrimos(context.Background(), maxPrimos)
		duracionComputo = time.Since(inicioComputo)
	}

	duracionTotal := time.Since(inicioTotal)
	
	fmt.Printf("[Secuencial] Decisión (%s): %d. Rama: %s. Tpo Dec: %v, Tpo Cómputo: %v, Tpo Total: %v\n",
		"Traza > Umbral", traza, ramaSeleccionada, duracionDecision, duracionComputo, duracionTotal)
		
	return duracionTotal, ramaSeleccionada
}

// =========================================================================
// PASO 4: IMPLEMENTACIÓN ESPECULATIVA (Concurrente)
// =========================================================================

func ejecucionEspeculativa(n, umbral, dificultadPoW, maxPrimos int, blockData string) (time.Duration, string, error) {
	inicioTotal := time.Now()

	// 1. Crear el contexto para la CANCELACIÓN
	ctx, cancel := context.WithCancel(context.Background())
	// Defer cancel() asegura que todas las goroutines (incluso la ganadora)
	// reciban la señal de cancelación al final, limpiando recursos.
	defer cancel()

	// 2. Crear los channels para los resultados
	chRamaA := make(chan ResultadoPoW, 1)      // Buffer 1 para que la goroutine no se bloquee
	chRamaB := make(chan ResultadoPrimos, 1)   // Buffer 1
	chDecision := make(chan ResultadoDecision, 1) // Buffer 1

	// 3. Lanzar Goroutines

	// Goroutine 1: Rama A (ProofOfWork)
	go func() {
		inicioRamaA := time.Now()
		hash, nonce, err := SimularProofOfWork(ctx, blockData, dificultadPoW) // Pasamos el contexto cancelable
		duracionRamaA := time.Since(inicioRamaA)
		chRamaA <- ResultadoPoW{hash: hash, nonce: nonce, duracion: duracionRamaA, err: err}
	}()

	// Goroutine 2: Rama B (EncontrarPrimos)
	go func() {
		inicioRamaB := time.Now()
		primos, err := EncontrarPrimos(ctx, maxPrimos) // Pasamos el contexto cancelable
		duracionRamaB := time.Since(inicioRamaB)
		chRamaB <- ResultadoPrimos{primos: primos, duracion: duracionRamaB, err: err}
	}()

	// Goroutine 3: Decisión (CalcularTraza)
	go func() {
		inicioDecision := time.Now()
		traza := CalcularTrazaDeProductoDeMatrices(n)
		duracionDecision := time.Since(inicioDecision)
		chDecision <- ResultadoDecision{traza: traza, duracion: duracionDecision}
	}()

	// 4. Sincronización, Selección y Cancelación
	
	// Primero, esperamos SÓLO el resultado de la decisión
	resDecision := <-chDecision
	
	var ramaSeleccionada string
	var err error
	
	// 5. Tomar la decisión
	if resDecision.traza > umbral {
		ramaSeleccionada = "A (ProofOfWork)"
		// Esperamos el resultado de la rama A
		resA := <-chRamaA
		if resA.err != nil {
			// Esto podría ser un error de cancelación si la decisión fue muy rápida
			// o un error real de la función.
			err = resA.err
		}
		// Inmediatamente después de recibir el resultado de A,
		// cancelamos el contexto. Esto detendrá a la Goroutine B.
		cancel() 
		
	} else {
		ramaSeleccionada = "B (EncontrarPrimos)"
		// Esperamos el resultado de la rama B
		resB := <-chRamaB
		if resB.err != nil {
			err = resB.err
		}
		// Cancelamos el contexto. Esto detendrá a la Goroutine A.
		cancel()
	}

	duracionTotal := time.Since(inicioTotal)

	fmt.Printf("[Especulativa] Decisión (%s): %d. Rama: %s. Tpo Dec: %v, Tpo Total: %v\n",
		"Traza > Umbral", resDecision.traza, ramaSeleccionada, resDecision.duracion, duracionTotal)

	// Manejamos el caso donde la rama ganadora también fue cancelada (si la decisión fue MUY rápida)
	if err != nil && strings.Contains(err.Error(), "cancelado") {
		return 0, ramaSeleccionada, fmt.Errorf("la rama ganadora fue cancelada (la decisión fue más rápida)")
	}
	
	return duracionTotal, ramaSeleccionada, nil
}

// =========================================================================
// PASO 5: FUNCIÓN MAIN (Orquestador y Analista)
// =========================================================================

func main() {
	// 1. Manejo de Parámetros por línea de comandos
	// go run main.go [n] [umbral] [dificultadPoW] [maxPrimos] [nombre_archivo]
	if len(os.Args) != 6 {
		fmt.Println("Uso: go run main.go [n] [umbral] [dificultadPoW] [maxPrimos] [nombre_archivo]")
		return
	}

	n, _ := strconv.Atoi(os.Args[1])
	umbral, _ := strconv.Atoi(os.Args[2])
	dificultadPoW, _ := strconv.Atoi(os.Args[3])
	maxPrimos, _ := strconv.Atoi(os.Args[4])
	nombreArchivo := os.Args[5]
	blockData := "datos_del_bloque" // Data fija para PoW

	fmt.Printf("Iniciando simulación con:\n")
	fmt.Printf("  n (matrices): %d\n", n)
	fmt.Printf("  umbral (traza): %d\n", umbral)
	fmt.Printf("  dificultad (PoW): %d\n", dificultadPoW)
	fmt.Printf("  max (Primos): %d\n", maxPrimos)
	fmt.Printf("  archivo: %s\n\n", nombreArchivo)


	// 2. Análisis de Rendimiento (30 ejecuciones)
	const numEjecuciones = 30
	var tiemposSecuenciales []time.Duration
	var tiemposEspeculativos []time.Duration

	fmt.Println("--- INICIANDO EJECUCIÓN SECUENCIAL ---")
	for i := 0; i < numEjecuciones; i++ {
		duracion, _ := ejecucionSecuencial(n, umbral, dificultadPoW, maxPrimos, blockData)
		tiemposSecuenciales = append(tiemposSecuenciales, duracion)
		time.Sleep(100 * time.Millisecond) // Pequeña pausa
	}
	fmt.Println("\n--- INICIANDO EJECUCIÓN ESPECULATIVA ---")
	for i := 0; i < numEjecuciones; i++ {
		duracion, _, err := ejecucionEspeculativa(n, umbral, dificultadPoW, maxPrimos, blockData)
		if err != nil {
			fmt.Printf("Error en ejecución especulativa: %v\n", err)
		} else {
			tiemposEspeculativos = append(tiemposEspeculativos, duracion)
		}
		time.Sleep(100 * time.Millisecond) // Pequeña pausa
	}

	// 3. Calcular Promedios y Speedup
	var totalSecuencial time.Duration
	for _, t := range tiemposSecuenciales {
		totalSecuencial += t
	}
	tpoPromedioSecuencial := totalSecuencial / time.Duration(len(tiemposSecuenciales))

	var totalEspeculativo time.Duration
	for _, t := range tiemposEspeculativos {
		totalEspeculativo += t
	}
	tpoPromedioEspeculativo := totalEspeculativo / time.Duration(len(tiemposEspeculativos))

	speedup := float64(tpoPromedioSecuencial) / float64(tpoPromedioEspeculativo)

	// 4. Reporte Final (Consola y Archivo)
	
	// Crear el contenido del reporte
	reporte := fmt.Sprintf("--- ANÁLISIS DE RENDIMIENTO ---\n")
	reporte += fmt.Sprintf("Parámetros: n=%d, umbral=%d, dificultadPoW=%d, maxPrimos=%d\n", n, umbral, dificultadPoW, maxPrimos)
	reporte += fmt.Sprintf("---------------------------------\n")
	reporte += fmt.Sprintf("Tiempo Promedio Secuencial:   %v\n", tpoPromedioSecuencial)
	reporte += fmt.Sprintf("Tiempo Promedio Especulativo: %v\n", tpoPromedioEspeculativo)
	reporte += fmt.Sprintf("---------------------------------\n")
	reporte += fmt.Sprintf("Speedup (Secuencial / Especulativo): %.2f\n", speedup)

	fmt.Println("\n" + reporte)

	// Escribir en el archivo
	err := os.WriteFile(nombreArchivo, []byte(reporte), 0644)
	if err != nil {
		fmt.Printf("Error al escribir el archivo de reporte: %v\n", err)
	} else {
		fmt.Printf("Reporte guardado exitosamente en '%s'\n", nombreArchivo)
	}
}