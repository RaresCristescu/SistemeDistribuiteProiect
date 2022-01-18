// Nume: Blaj Ioan-David, Cristescu Rares Dorian
// Grupa: 454, 454
// Rolul programului: retea de gossip SI (push, pull sau push-pull)

package main

import(
	"fmt"
	"time"
	"math/rand"
	"os"
	"bufio"
)

type Nod struct{
	id int
	mesaj string
}

func InsereazaNod(id int, mesaj string, n int, canale_requesturi []chan int, canale_actualizari []chan string, noduri_infectate chan int, start_ciclu chan int, log chan string, mesaje_trimise chan int, push int, pull int){
	nod := Nod{id,mesaj}
	rand.Seed(time.Now().Unix())
	var index_vecin, vecin_id, ciclu int
	for{
		ciclu=<-start_ciclu
		if ciclu<1 || ciclu>n{
			fmt.Printf("Numarul ciclului este anormal(%d)!\n",ciclu)
			panic("Opresc programul...")
		}
		// Aleg aleator un vecin
		index_vecin = rand.Intn(n-1) // Aleg aleator un vecin intre [0,n-2]
		copy_index_vecin:=index_vecin
		for i:=0; i<n; i++{
			if i!=nod.id{
				copy_index_vecin-=1
			}
			if copy_index_vecin==-1{
				vecin_id=i
				break
			}
		}
		// Push - Daca nodul este infectat, trimite actualizarea la vecin
		if push==1{
			if nod.mesaj!=""{
				canale_actualizari[vecin_id]<-nod.mesaj
				mesaje_trimise<-1
			}
		}
		// Pull - Daca nodul nu este infectat face request de actualizare la vecin
		if pull==1{
			if nod.mesaj==""{
				canale_requesturi[vecin_id]<-nod.id
				mesaje_trimise<-1
			}
		}
		
		time.Sleep(50*time.Millisecond)
		// Push - Verifica daca exista o infectie care vine spre mine si ia actiunile necesare
		if push==1{
			if nod.mesaj==""{
				if len(canale_actualizari[nod.id])>0{
					nod.mesaj=<-canale_actualizari[nod.id]
					log<-fmt.Sprintf("Nodul %d a fost infectat cu mesajul \"%s\"(ciclul %d)\n", nod.id, nod.mesaj, ciclu)
					noduri_infectate<-1
				}
			}
			for len(canale_actualizari[nod.id])>0{
				<-canale_actualizari[nod.id]
			}
		}
		// Pull - Verifica daca exista un request destinat mie pentru a trimite actualizarea si ia actiunile necesare
		if pull==1{
			if nod.mesaj!=""{
				for len(canale_requesturi[nod.id])>0{
					destinatar_id:=<-canale_requesturi[nod.id]
					canale_actualizari[destinatar_id]<-nod.mesaj
					mesaje_trimise<-1
					//noduri_infectate<-1
					//log<-fmt.Sprintf("Nodul %d a fost infectat cu mesajul \"%s\" de catre nodul %d(ciclul %d)\n", destinatar_id, nod.mesaj, nod.id, ciclu)
				}
			}
			// Daca am request-uri, dar nu sunt infectat, decartez request-urile
			for len(canale_requesturi[nod.id])>0{
				<-canale_requesturi[nod.id]
			}
		}
		
		time.Sleep(50*time.Millisecond)
		// Daca am facut pull, vad daca am primit actualizarea
		if pull==1{
			if nod.mesaj=="" && len(canale_actualizari[nod.id])>0{
				nod.mesaj=<-canale_actualizari[nod.id]
				noduri_infectate<-1
				log<-fmt.Sprintf("Nodul %d a fost infectat cu mesajul \"%s\"(ciclul %d)\n", nod.id, nod.mesaj, ciclu)
			}
		}
	}
}

func main(){
	var n,ciclu,mesaje,optiune,push,pull int=0,0,0,0,0,0
	var mesaj string=""
	fmt.Printf("Introdu optiunea (1=push, 2=pull, 3=push-pull): ")
	if _,err:=fmt.Scan(&optiune); err!=nil{
    	panic(err)
	}
	if optiune!=1 && optiune!=2 && optiune!=3{
		panic("Optiune invalida!")
	}else{
		if optiune==1{
			push=1
			pull=0
		}else if optiune==2{
			push=0
			pull=1
		}else{
			push=1
			pull=1
		}
	}
	fmt.Printf("Introdu numarul de noduri din retea: ")
	if _,err:=fmt.Scan(&n); err!=nil{
    	panic(err)
	}
	fmt.Printf("Introdu mesajul: ")
	if _,err:=fmt.Scan(&mesaj); err!=nil{
    	panic(err)
	}
	
	var canale_requesturi=make([]chan int, n)
	var canale_actualizari=make([]chan string, n)
	noduri_infectate:=make(chan int, n-1) // maxim n-1 noduri pot fi infectate in mod nou intr-un ciclu
	mesaje_trimise:=make(chan int, n+n-1) // maxim n+n-1 mesaje pot fi trimise intr-un ciclu
	start_ciclu:=make(chan int, n)
	log:=make(chan string, n-1) // mesajele tip log anunta cand un nod s-a infectat (maxim n-1 intr-un ciclu)
	fmt.Printf("Generez reteaua de %d noduri...\n",n)
	
	// Logging
	log_file,err:=os.Create("log.txt")
	if err!=nil{
		panic(err)
	}
	w:=bufio.NewWriter(log_file)
	for i:=0; i<n; i++{
		canale_requesturi[i]=make(chan int, n-1) // un nod poate primi maxim n-1 requesturi intr-un ciclu
		canale_actualizari[i]=make(chan string, n-1+1) // un nod poate primi maxim n-1+1 actualizari intr-un ciclu (fanout=1)
		if i!=n-1{
			go InsereazaNod(i,"",n,canale_requesturi,canale_actualizari,noduri_infectate,start_ciclu,log,mesaje_trimise,push,pull)
		}else{
			fmt.Printf("Infectez nodul %d cu mesajul \"%s\"\n",i,mesaj)
			_,err=fmt.Fprintf(w, "Infectez nodul %d cu mesajul \"%s\"(ciclul %d)\n",i,mesaj,ciclu)
			if err!=nil{
				panic(err)
			}
			go InsereazaNod(i,mesaj,n,canale_requesturi,canale_actualizari,noduri_infectate,start_ciclu,log,mesaje_trimise,push,pull)
		}
	}

	aux_n:=n-1 // Ultimul nod este deja infectat
	// Start timer
	start := time.Now()
	for{
		ciclu+=1
		for i:=0; i<n; i++{
			start_ciclu<-ciclu
		}

		time.Sleep(200*time.Millisecond)

		for len(log)>0{
			_,err=fmt.Fprint(w, <-log)
			if err!=nil{
				panic(err)
			}
		}
		w.Flush()
		for len(mesaje_trimise)>0{
			if confirmare:=<-mesaje_trimise; confirmare==1{
				mesaje+=1
			}
		}
		for len(noduri_infectate)>0{
			if confirmare:=<-noduri_infectate; confirmare==1{
				aux_n-=1
				if aux_n==0{
					timp_acum:=time.Now()
					fmt.Printf("Toate nodurile au fost infectate in %dms(%d cicluri).\n",timp_acum.Sub(start)/1000000,ciclu)
					fmt.Printf("Au fost trimise %d mesaje.\n", mesaje)
					os.Exit(0);
				}
			}
		}
	}
}