# Grid Route Optimizer — Copilot Instructions

## 1. Cel projektu

Projekt jest aplikacją webową służącą do **wyznaczania, optymalizacji i wizualizacji zbiorów tras pokrywających wybrane podgrafy krat**.

Głównym zastosowaniem jest planowanie tras przechodzących przez wybrane wierzchołki regularnej kraty.

System powinien umożliwiać użytkownikowi:

* tworzenie kraty,
* określanie jej rozmiaru,
* interaktywne poruszanie się po kracie,
* powiększanie i pomniejszanie widoku,
* przesuwanie planszy,
* zaznaczanie wierzchołków,
* edycję połączeń pomiędzy wierzchołkami,
* dodawanie i usuwanie ścian/przeszkód,
* definiowanie obszaru o dowolnym kształcie,
* określanie wierzchołków należących do wybranego obszaru,
* definiowanie ograniczeń dotyczących tras,
* uruchamianie algorytmów optymalizacyjnych,
* wizualizowanie znalezionych tras,
* analizowanie jakości rozwiązania,
* prezentowanie statystyk,
* porównywanie różnych rozwiązań i algorytmów.

W przyszłości system powinien zostać rozszerzony o możliwość pracy z **trójwymiarowymi kratami**.

Projekt jest pracą dyplomową, dlatego szczególnie ważne są:

* czytelna architektura,
* modularność,
* testowalność,
* możliwość przeprowadzania eksperymentów,
* możliwość porównywania algorytmów,
* mierzalność wyników,
* łatwość rozszerzania systemu,
* oddzielenie logiki biznesowej od wizualizacji.

---

# 2. Główne założenia architektury

Aplikacja jest systemem klient-serwer.

```text
┌─────────────────────────────────────────────┐
│                  FRONTEND                   │
│                                             │
│ React + TypeScript                          │
│ Vite                                        │
│ PixiJS                                      │
│ Three.js (przyszłe 3D)                     │
│                                             │
│ UI + interakcja + wizualizacja              │
└──────────────────────┬──────────────────────┘
                       │
                       │ HTTP / REST / JSON
                       │
┌──────────────────────▼──────────────────────┐
│                  BACKEND                    │
│                                             │
│ Go                                          │
│                                             │
│ Model grafu                                 │
│ Generowanie krat                            │
│ Podgrafy                                    │
│ Trasy                                       │
│ Algorytmy optymalizacji                     │
│ Ewaluacja rozwiązań                         │
│                                             │
└─────────────────────────────────────────────┘
```

## Najważniejsza zasada

**Frontend nie powinien zawierać właściwej logiki algorytmów optymalizacyjnych.**

Frontend odpowiada przede wszystkim za:

* prezentację danych,
* interakcję użytkownika,
* wizualizację,
* obsługę stanu UI,
* komunikację z backendem.

Backend odpowiada za:

* model grafu,
* generowanie krat,
* reprezentację podgrafów,
* walidację danych,
* algorytmy,
* optymalizację,
* obliczanie statystyk,
* ocenę rozwiązań.

Algorytmy optymalizacyjne powinny być implementowane w Go.

Nie należy przenosić właściwych algorytmów optymalizacyjnych do TypeScript tylko dlatego, że ich wynik jest wizualizowany w przeglądarce.

---

# 3. Technologie

## Frontend

Podstawowe technologie:

* React
* TypeScript
* Vite
* PixiJS
* Three.js — przyszłe rozszerzenie 3D

React odpowiada za:

* strukturę aplikacji,
* UI,
* komponenty,
* stan interfejsu.

PixiJS odpowiada za:

* wydajne renderowanie planszy 2D,
* renderowanie dużych krat,
* elementy graficzne planszy,
* interakcję z obszarem roboczym.

Three.js będzie wykorzystany w przyszłości do:

* renderowania krat 3D,
* obsługi kamery 3D,
* wizualizacji tras 3D.

---

## Backend

Podstawową technologią backendu jest:

* Go

Backend powinien w miarę możliwości korzystać ze standardowej biblioteki Go.

Dodatkowe biblioteki należy dodawać tylko wtedy, gdy dają wyraźną korzyść.

API powinno być REST API wykorzystującym JSON.

---

# 4. Struktura projektu

Docelowa struktura powinna przypominać:

```text
grid-route-optimizer/
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── features/
│   │   ├── graph/
│   │   ├── rendering/
│   │   ├── api/
│   │   ├── types/
│   │   ├── hooks/
│   │   ├── state/
│   │   ├── utils/
│   │   ├── App.tsx
│   │   └── main.tsx
│   │
│   ├── package.json
│   └── ...
│
├── backend/
│   ├── cmd/
│   ├── internal/
│   │   ├── graph/
│   │   ├── grid/
│   │   ├── subgraph/
│   │   ├── routes/
│   │   ├── optimization/
│   │   ├── evaluation/
│   │   └── api/
│   │
│   ├── tests/
│   ├── go.mod
│   └── ...
│
├── docker-compose.yml
├── .gitignore
├── README.md
└── copilot_instructions.md
```

Nie należy tworzyć wszystkich katalogów od razu.

Struktura powinna rozwijać się wraz z projektem.

Nie należy tworzyć nieużywanych abstrakcji wyłącznie „na przyszłość”.

---

# 5. Model domenowy

Najważniejszym założeniem projektu jest rozdzielenie:

```text
MODEL GRAFU
```

od:

```text
WIZUALIZACJI GRAFU
```

Model grafu nie może zależeć od React, PixiJS ani Three.js.

Backend powinien posiadać własny model grafu.

Frontend może posiadać odpowiedniki typów potrzebnych do komunikacji z backendem, ale nie powinien definiować logiki domenowej niezależnej od backendu.

---

# 6. Graf

Podstawowym modelem matematycznym jest graf:

```text
G = (V, E)
```

gdzie:

* `V` — zbiór wierzchołków,
* `E` — zbiór krawędzi.

W kontekście aplikacji graf jest wizualizowany jako **plansza złożona z pól**.

---

# 7. Najważniejsze założenie wizualizacji — pola zamiast punktów

Wizualizacja grafu **nie jest klasycznym diagramem grafowym**.

Wierzchołki nie powinny być przedstawiane jako punkty/kropki połączone liniami.

**Każdy wierzchołek grafu jest reprezentowany jako kwadrat/pole na planszy.**

Przykład:

```text
┌───┐ ┌───┐ ┌───┐
│   │ │   │ │   │
└───┘ └───┘ └───┘
```

Każdy kwadrat odpowiada jednemu wierzchołkowi.

Użytkownik powinien postrzegać aplikację bardziej jako:

* planszę,
* siatkę,
* środowisko,
* mapę,

niż klasyczny diagram grafowy.

---

# 8. Połączenia między polami

Dwa sąsiednie pola są domyślnie połączone, jeżeli można bezpośrednio przejść z jednego pola do drugiego.

Przykład:

```text
┌───┐┌───┐
│ A ││ B │
└───┘└───┘
```

oznacza:

```text
A ─ B
```

czyli istnieje krawędź pomiędzy A i B.

Połączenie jest więc związane z **przestrzenią pomiędzy polami**.

---

# 9. Brak połączenia — pusta przestrzeń

Jeżeli pomiędzy dwoma polami znajduje się pusta przestrzeń/przerwa, pola nie są połączone.

Przykład:

```text
┌───┐       ┌───┐
│ A │       │ B │
└───┘       └───┘
```

oznacza:

```text
A    B
```

bez krawędzi.

Pusta przestrzeń pomiędzy polami ma znaczenie semantyczne i oznacza brak bezpośredniego przejścia.

---

# 10. Brak połączenia — ściana

Brak połączenia może również zostać przedstawiony za pomocą grubszej ściany.

Przykład:

```text
┌───┐███┌───┐
│ A │███│ B │
└───┘███└───┘
```

oznacza:

```text
A    B
```

czyli brak krawędzi.

Ściana reprezentuje fizyczną przeszkodę uniemożliwiającą przejście.

---

# 11. Zwykła granica pola a ściana

Należy wyraźnie rozróżnić:

* zwykłą granicę pomiędzy sąsiednimi polami,
* ścianę oznaczającą brak połączenia.

Przykład zwykłego połączenia:

```text
┌─────┬─────┐
│  A  │  B  │
└─────┴─────┘
```

A i B są połączone.

Przykład ściany:

```text
┌─────┐███┌─────┐
│  A  │███│  B  │
└─────┘███└─────┘
```

A i B nie są połączone.

Wizualizacja musi umożliwiać jednoznaczne rozpoznanie tej różnicy.

---

# 12. Sąsiedztwo w kracie 2D

Standardowa krata 2D korzysta z sąsiedztwa:

```text
      N
      │
      ▼
  W ← X → E
      ▲
      │
      S
```

czyli:

* góra,
* dół,
* lewo,
* prawo.

Domyślnie pola po przekątnej nie są połączone.

```text
X       X
  \   /
   \ /
   / \
  /   \
X       X
```

Połączenia diagonalne mogą zostać dodane w przyszłości jako jawna konfiguracja grafu, ale nie powinny być domyślne.

---

# 13. Krawędź jako możliwość przejścia

Krawędź nie powinna być rozumiana wyłącznie jako linia narysowana na ekranie.

W modelu domenowym:

```text
Edge(A, B)
```

oznacza możliwość przejścia pomiędzy dwoma wierzchołkami.

Wizualnie może być ona reprezentowana przez brak przeszkody pomiędzy dwoma sąsiednimi polami.

Dlatego:

```text
┌───┬───┐
│ A │ B │
└───┴───┘
```

oznacza:

```text
Edge(A, B) = true
```

natomiast:

```text
┌───┐███┌───┐
│ A │███│ B │
└───┘███└───┘
```

oznacza:

```text
Edge(A, B) = false
```

---

# 14. Ściany są częścią modelu grafu

Ściana nie może być wyłącznie efektem wizualnym.

Jeżeli użytkownik doda ścianę pomiędzy dwoma polami:

```text
A ███ B
```

odpowiednia krawędź musi zostać usunięta albo oznaczona jako niedostępna w modelu domenowym.

Algorytm optymalizacyjny musi uwzględniać tę zmianę.

Frontend odpowiada za:

* wyświetlenie ściany,
* interakcję z użytkownikiem,
* wysłanie zmiany do backendu.

Backend jest źródłem prawdy dotyczącej grafu.

---

# 15. Edycja połączeń

Użytkownik powinien móc edytować przejścia pomiędzy sąsiednimi polami.

Przykład:

```text
┌───┬───┬───┐
│   │   │   │
├───┼───┼───┤
│   │   │   │
├───┼───┼───┤
│   │   │   │
└───┴───┴───┘
```

Po dodaniu ściany:

```text
┌───┬───┬───┐
│   │   │   │
├───┼███┼───┤
│   │   │   │
├───┼───┼───┤
│   │   │   │
└───┴───┴───┘
```

odpowiednie przejście staje się niedostępne.

Zmiana wizualna musi odpowiadać zmianie modelu grafu.

---

# 16. Krata 2D

Krata 2D jest specjalnym rodzajem grafu.

Przykład:

```text
┌───┬───┬───┬───┐
│   │   │   │   │
├───┼───┼───┼───┤
│   │   │   │   │
├───┼───┼───┼───┤
│   │   │   │   │
├───┼───┼───┼───┤
│   │   │   │   │
└───┴───┴───┴───┘
```

Każde pole reprezentuje wierzchołek.

Domyślnie sąsiednie pola są połączone.

---

# 17. Model wierzchołka

Wierzchołek powinien posiadać co najmniej:

```text
ID
X
Y
```

W przyszłości model powinien umożliwiać:

```text
Z
```

oraz opcjonalnie:

```text
Weight
Type
Cost
Metadata
```

Nie należy uzależniać modelu wierzchołka od sposobu jego renderowania.

---

# 18. Model krawędzi

Krawędź powinna posiadać informacje wystarczające do określenia:

```text
From
To
Weight
Available / Blocked
```

W przyszłości możliwe jest dodanie:

```text
Direction
Cost
Type
Metadata
```

---

# 19. Model kraty a model renderowania

Model domenowy:

```text
Grid
  ├── Vertices
  └── Edges
```

nie może znać:

```text
Pixi.Graphics
PIXI.Container
Three.Mesh
React Component
```

Renderowanie jest osobną warstwą.

---

# 20. Plansza

Plansza jest głównym elementem interfejsu.

Powinna być podobna w zachowaniu do mapy lub edytora przestrzennego.

Przykładowy układ:

```text
┌───────────────────────────────────────────────────────────────┐
│                         TOP BAR                               │
│ Grid Route Optimizer      [New] [Save] [Optimize]             │
├────────────────┬──────────────────────────────────────────────┤
│                │                                              │
│   SIDE PANEL   │                  GRID CANVAS                 │
│                │                                              │
│ Configuration  │                                              │
│                │             ┌───┬───┬───┐                   │
│ Grid           │             │   │   │   │                   │
│ Selection      │             ├───┼███┼───┤                   │
│ Routes         │             │   │   │   │                   │
│ Optimization   │             └───┴───┴───┘                   │
│                │                                              │
│ Statistics     │                                              │
│                │                                              │
├────────────────┴──────────────────────────────────────────────┤
│                         STATUS BAR                            │
└───────────────────────────────────────────────────────────────┘
```

UI wokół planszy powinno być niezależne od przesuwania i zoomowania planszy.

---

# 21. Pan

Użytkownik musi móc przesuwać planszę.

Pan powinien:

* działać płynnie,
* nie przesuwać całej strony,
* zmieniać pozycję kamery/widoku,
* działać niezależnie od modelu grafu.

Preferowany model:

```text
Grid coordinates
        ↓
Camera transform
        ↓
Screen coordinates
```

Nie należy modyfikować współrzędnych wierzchołków tylko dlatego, że użytkownik przesunął widok.

---

# 22. Zoom

Użytkownik musi móc:

* zoomować kółkiem myszy,
* korzystać z przycisków `+` i `-`,
* w przyszłości korzystać z gestów dotykowych.

Zoom powinien być płynny.

Preferowane jest zoomowanie względem pozycji kursora.

Przykładowe narzędzia:

```text
[ + ] [ - ] [ Fit ] [ Reset ]
```

---

# 23. Reset widoku

Plansza powinna umożliwiać:

* reset zoomu,
* reset pozycji,
* wycentrowanie kraty,
* dopasowanie całej kraty do dostępnego obszaru.

---

# 24. Renderowanie 2D

Do renderowania planszy 2D należy używać PixiJS.

Nie należy tworzyć osobnego elementu DOM/React dla każdego pola przy dużych instancjach grafu.

Plansza powinna być renderowana wydajnie z wykorzystaniem możliwości GPU.

React powinien zarządzać aplikacją i UI, natomiast PixiJS powinien zarządzać właściwym renderowaniem planszy.

---

# 25. Warstwy wizualizacji

Plansza powinna być logicznie podzielona na warstwy:

```text
Grid background
      ↓
Edges / passages
      ↓
Vertices / cells
      ↓
Walls
      ↓
Subgraph
      ↓
Target vertices
      ↓
Routes
      ↓
Selection
      ↓
Interaction overlay
```

Poszczególne warstwy powinny być możliwe do niezależnego aktualizowania i w przyszłości włączania/wyłączania.

---

# 26. Stany wizualne pól

Pole może znajdować się w różnych stanach:

1. zwykłe,
2. zaznaczone,
3. należące do podgrafu,
4. należące do obszaru docelowego,
5. odwiedzane przez trasę,
6. należące do wielu tras,
7. aktualnie wskazane kursorem.

Stany powinny być reprezentowane spójnie wizualnie.

---

# 27. Stany wizualne przejść

Przejście może być:

1. dostępne,
2. niedostępne,
3. należące do trasy,
4. należące do wielu tras,
5. aktualnie edytowane,
6. zaznaczone.

Ściana powinna być jednoznacznie odróżnialna od zwykłej granicy pola.

---

# 28. Definiowanie obszaru

Jedną z głównych funkcjonalności jest możliwość zdefiniowania **obszaru o dowolnym kształcie**.

Użytkownik może narysować wielokąt:

```text
        ●────────●
       /          \
      /            \
     ●              ●
      \            /
       \──────●───/
```

System określa następnie, które pola/wierzchołki kraty znajdują się wewnątrz tego obszaru.

---

# 29. SelectionRegion i Subgraph

Należy rozróżnić:

```text
SelectionRegion
```

od:

```text
Subgraph
```

`SelectionRegion` jest obiektem geometrycznym określonym przez użytkownika.

`Subgraph` jest wynikiem przekształcenia obszaru geometrycznego w zbiór wierzchołków/krawędzi grafu.

Przykładowa zależność:

```text
Grid
  │
  ▼
SelectionRegion
  │
  ▼
Subgraph
  │
  ▼
Target vertices
```

Nie należy mieszać tych pojęć.

---

# 30. Zaznaczanie

Frontend powinien umożliwiać różne tryby pracy:

```text
Select
Pan
Region
Edit
```

Aktualny tryb powinien być widoczny dla użytkownika.

Możliwe jest późniejsze dodanie kolejnych narzędzi.

---

# 31. Trasy

Trasa jest niezależnym obiektem domenowym.

Przykładowy model:

```text
Route
 ├── Vertices
 ├── Edges
 ├── Length
 └── Metadata
```

Trasa nie jest tylko efektem wizualnym.

Backend powinien zwracać rzeczywiste trasy, a frontend powinien je wizualizować.

---

# 32. Wizualizacja tras

Trasa powinna być widoczna bezpośrednio na planszy.

Możliwe sposoby:

* wyróżnianie pól,
* wyróżnianie przejść,
* strzałki,
* numerowanie kolejności odwiedzin,
* oznaczenie początku i końca.

Szczegóły wizualizacji mogą ewoluować, ale znaczenie trasy musi pozostać zgodne z modelem backendu.

---

# 33. Nakładanie się tras

Jeżeli kilka tras przechodzi przez ten sam wierzchołek lub krawędź, wizualizacja powinna umożliwiać rozpoznanie tego faktu.

W przyszłości można wykorzystać:

* różne oznaczenia,
* grubość przejścia,
* licznik wykorzystania,
* osobne warstwy tras.

Nie należy jednak zakładać jednego konkretnego sposobu wizualizacji na zawsze.

---

# 34. Model podgrafu

Podgraf powinien być reprezentowany niezależnie od wizualnego obszaru.

Przykładowo:

```text
Subgraph
 ├── Vertices
 └── Edges
```

W zależności od definicji problemu podgraf może zawierać:

* tylko wybrane wierzchołki,
* wybrane wierzchołki i istniejące pomiędzy nimi krawędzie,
* dodatkowe informacje dotyczące obszaru.

---

# 35. Problem optymalizacyjny

Głównym problemem jest znalezienie zbioru tras, które łącznie pokrywają wybrane wierzchołki.

W uproszczeniu:

```text
V_target ⊆ V(route1) ∪ ... ∪ V(routek)
```

przy jednoczesnej minimalizacji:

```text
L(route1) + ... + L(routek)
```

gdzie `L(route)` oznacza długość trasy.

Dokładna funkcja celu może zostać rozszerzona w zależności od przyjętego algorytmu.

---

# 36. Ograniczenia

System powinien umożliwiać wprowadzanie ograniczeń.

Potencjalne ograniczenia:

* maksymalna liczba tras,
* dokładna liczba tras,
* maksymalna długość pojedynczej trasy,
* maksymalna łączna długość tras,
* wymagany początek,
* wymagany koniec,
* określone punkty startowe,
* określone punkty końcowe,
* niedostępne wierzchołki,
* niedostępne krawędzie,
* ściany/przeszkody.

Nie wszystkie ograniczenia muszą być dostępne od początku.

Architektura powinna umożliwiać ich późniejsze dodawanie.

---

# 37. Algorytmy optymalizacyjne

Algorytmy powinny znajdować się w backendzie.

Nie powinny być zależne od:

* HTTP,
* JSON,
* React,
* PixiJS,
* Three.js.

Preferowana architektura:

```text
HTTP Handler
      ↓
Optimization Service
      ↓
Optimization Problem
      ↓
Optimizer
      ↓
Solution
```

---

# 38. Wiele algorytmów

Projekt powinien być przygotowany na implementację wielu algorytmów.

Przykładowo:

```text
Optimizer
 ├── Algorithm A
 ├── Algorithm B
 └── Algorithm C
```

Każdy algorytm powinien otrzymywać problem optymalizacyjny i zwracać rozwiązanie.

Pozwoli to później porównywać różne metody.

---

# 39. Rozwiązanie

Wynik optymalizacji powinien być reprezentowany jako:

```text
Solution
 ├── Routes
 ├── TotalLength
 ├── NumberOfRoutes
 ├── CoveredVertices
 └── Metrics
```

W przyszłości można dodać:

```text
ExecutionTime
Algorithm
Iterations
OptimalityGap
CoverageRatio
```

---

# 40. Ewaluacja

System powinien umożliwiać ocenę jakości rozwiązania.

Podstawowe metryki:

```text
Total route length
Number of routes
Number of covered vertices
Coverage ratio
Execution time
Number of repeated vertices
Number of repeated edges
```

W przyszłości można dodać kolejne metryki.

---

# 41. Statystyki w UI

Po zakończeniu optymalizacji użytkownik powinien zobaczyć podsumowanie:

```text
Solution

Routes:          5
Total length:    183.5
Covered nodes:   240
Coverage:        100%
Execution time:  1.42 s
```

Statystyki mogą znajdować się w bocznym panelu lub osobnym panelu wyników.

---

# 42. Porównywanie algorytmów

W przyszłości aplikacja powinna umożliwiać uruchomienie różnych algorytmów dla tego samego problemu.

Przykład:

```text
                 A       B       C
--------------------------------------
Length          120     115     118
Routes            5       5       4
Time           0.2s    1.4s    0.7s
Coverage        100%    100%    100%
```

Architektura backendu powinna umożliwiać takie porównania.

---

# 43. Stan aplikacji frontend

Stan powinien być logicznie rozdzielony.

## Stan problemu

```text
Grid
Subgraph
Target vertices
Walls
Constraints
```

## Stan rozwiązania

```text
Routes
Total length
Number of routes
Metrics
```

## Stan wizualizacji

```text
Zoom
Camera position
Pan offset
Selected vertex
Selected edge
Selected tool
Visible layers
```

Nie należy mieszać stanu wizualizacji z modelem matematycznym.

Przykład:

```text
zoom = 1.5
```

nie jest właściwością grafu.

---

# 44. API frontend

Komunikacja z backendem powinna być skupiona w osobnej warstwie.

Preferowana struktura:

```text
src/
└── api/
    ├── client.ts
    ├── gridApi.ts
    ├── subgraphApi.ts
    └── optimizationApi.ts
```

Komponenty React nie powinny bezpośrednio wykonywać wielu niezależnych `fetch()`.

Zamiast:

```text
Component
   ↓
fetch(...)
```

preferowane:

```text
Component
   ↓
API function
   ↓
HTTP client
   ↓
Backend
```

---

# 45. API backend

API powinno być REST API.

Przykładowe endpointy:

```text
GET  /api/health

POST /api/grids
GET  /api/grids/{id}

POST /api/subgraphs

POST /api/optimize

GET  /api/solutions/{id}
```

Dokładny kształt API może się zmieniać.

Logika domenowa nie powinna znajdować się bezpośrednio w handlerach HTTP.

Preferowany przepływ:

```text
HTTP Handler
    ↓
Request DTO
    ↓
Service
    ↓
Domain Model
    ↓
Algorithm
    ↓
Solution
    ↓
Response DTO
```

---

# 46. DTO a model domenowy

Modele API nie powinny być automatycznie utożsamiane z modelami domenowymi.

Przykład:

```text
OptimizationRequest
```

jest modelem API.

Natomiast:

```text
OptimizationProblem
```

jest modelem domenowym.

Należy zachować możliwość ich rozdzielenia.

---

# 47. Walidacja

Backend jest źródłem prawdy dla poprawności danych.

Frontend może wykonywać walidację dla wygody użytkownika, ale backend również musi sprawdzać:

* rozmiar kraty,
* poprawność indeksów,
* istnienie wierzchołków,
* poprawność krawędzi,
* poprawność ścian,
* poprawność parametrów,
* poprawność ograniczeń,
* poprawność tras,
* możliwość wykonania problemu.

Nie należy ufać danym otrzymanym z frontendu.

---

# 48. Wydajność

Wydajność jest ważnym elementem projektu.

Należy zakładać możliwość pracy z dużymi kratami:

```text
10 × 10
100 × 100
500 × 500
1000 × 1000
```

lub większymi.

Nie należy zakładać, że wszystkie instancje będą małe.

---

# 49. Wydajność frontendu

Należy unikać:

* tworzenia osobnego komponentu React dla każdego pola,
* niepotrzebnego renderowania całej planszy,
* przechowywania danych wizualnych i domenowych w jednym stanie,
* przesyłania całej kraty przez HTTP przy każdej zmianie,
* zbędnych aktualizacji PixiJS.

PixiJS powinien być wykorzystywany do wydajnego renderowania dużych plansz.

---

# 50. Culling i poziomy szczegółowości

Przy dużych grafach należy rozważyć:

* viewport culling,
* poziomy szczegółowości,
* pomijanie elementów niewidocznych,
* upraszczanie renderowania przy dużym oddaleniu,
* ograniczanie liczby aktualizowanych obiektów.

Przy dużym oddaleniu nie trzeba renderować wszystkich szczegółów pojedynczych pól, jeżeli nie są one widoczne.

---

# 51. Generowanie kraty

Jeżeli krata jest regularna i może być jednoznacznie opisana parametrami, frontend może generować część danych potrzebnych wyłącznie do wizualizacji lokalnie.

Nie należy jednak traktować lokalnie wygenerowanej wizualizacji jako źródła prawdy dla problemu optymalizacyjnego.

Backend pozostaje źródłem prawdy dla modelu domenowego.

---

# 52. 3D — przyszłe rozszerzenie

W przyszłości aplikacja może obsługiwać kraty 3D.

Model powinien być przygotowany na współrzędne:

```text
(x, y, z)
```

Docelowo:

```text
Grid2D
Grid3D
```

powinny korzystać ze wspólnych abstrakcji grafowych tam, gdzie ma to sens.

---

# 53. Renderowanie 3D

Docelowo:

```text
Graph / Grid
      │
      ├──────────► PixiJS
      │               ↓
      │              2D
      │
      └──────────► Three.js
                      ↓
                     3D
```

Model domenowy nie może zależeć od żadnego z rendererów.

---

# 54. Krata 3D

Wizualnie wierzchołki mogą być reprezentowane jako sześciany/pola przestrzenne.

Analogicznie do 2D:

* sąsiednie pola bez przeszkody → istnieje połączenie,
* przerwa → brak połączenia,
* ściana → brak połączenia.

Algorytm powinien operować na grafie, a nie na obiektach Three.js.

---

# 55. Interakcja 3D

W przyszłości użytkownik powinien móc:

* obracać kamerę,
* przesuwać kamerę,
* zoomować,
* wybierać wierzchołki,
* zaznaczać obszary,
* edytować połączenia,
* oglądać trasy,
* analizować rozwiązanie.

Nie należy implementować 3D przed ukończeniem podstawowej funkcjonalności 2D, chyba że jest to wymagane przez aktualne zadanie.

---

# 56. Animacja algorytmu

W przyszłości możliwe jest dodanie wizualizacji działania algorytmu krok po kroku.

Przykładowy model:

```text
OptimizationStep
```

może zawierać informacje o:

* aktualnym rozwiązaniu,
* aktualnie rozważanych wierzchołkach,
* aktualnych trasach,
* aktualnym koszcie,
* numerze iteracji.

Nie jest to wymagane w pierwszej wersji.

Architektura nie powinna jednak uniemożliwiać dodania tej funkcjonalności.

---

# 57. Eksperymenty

Projekt jest pracą dyplomową, dlatego ważne jest umożliwienie przeprowadzania eksperymentów.

W przyszłości można wprowadzić:

```text
Experiment
 ├── Problem
 ├── Algorithm
 ├── Parameters
 ├── Solution
 ├── Metrics
 └── ExecutionTime
```

Pozwoli to porównywać algorytmy na różnych instancjach problemu.

Nie należy dodawać bazy danych tylko dlatego, że może się przydać w przyszłości.

---

# 58. Docker

Docker jest przewidziany jako sposób uruchamiania całej aplikacji.

Docelowo:

```text
docker compose up
```

powinno umożliwiać uruchomienie:

```text
frontend
backend
```

Nie należy jednak komplikować lokalnego developmentu, jeżeli natywne uruchamianie jest prostsze.

---

# 59. Testy backendu

Backend powinien posiadać testy jednostkowe dla:

* generowania krat,
* grafów,
* wierzchołków,
* krawędzi,
* ścian,
* podgrafów,
* wyboru obszarów,
* obliczania długości tras,
* sprawdzania pokrycia,
* ograniczeń,
* algorytmów optymalizacyjnych.

Przykładowe przypadki:

```text
small grid
single vertex
empty target
single route
multiple routes
blocked edge
blocked vertex
unreachable vertex
maximum number of routes
```

---

# 60. Testy frontendu

Frontend powinien posiadać testy dla istotnej logiki:

* transformacji danych,
* obsługi stanu,
* wyboru narzędzi,
* zaznaczania,
* obliczania obszarów,
* komunikacji z API.

Testy renderowania PixiJS należy dodawać wtedy, gdy przynoszą rzeczywistą wartość.

---

# 61. Nazewnictwo

Kod powinien używać języka angielskiego.

Dotyczy to:

* nazw zmiennych,
* nazw funkcji,
* nazw typów,
* nazw plików,
* endpointów,
* komentarzy technicznych.

Przykłady:

```text
Grid
Grid2D
Grid3D
Vertex
Edge
Wall
Subgraph
SelectionRegion
TargetVertex
Route
Solution
OptimizationProblem
OptimizationResult
Optimizer
```

---

# 62. Komentarze

Kod powinien być czytelny sam z siebie.

Nie należy dodawać komentarzy opisujących oczywiste operacje.

Komentarze powinny wyjaśniać:

* dlaczego coś zostało zrobione,
* ograniczenia algorytmu,
* nietrywialne decyzje,
* założenia matematyczne,
* kwestie wydajnościowe.

Komentarze techniczne powinny być po angielsku.

---

# 63. Minimalna złożoność

Projekt powinien być rozwijany stopniowo.

Nie należy dodawać:

* mikroserwisów,
* skomplikowanych wzorców architektonicznych,
* dodatkowych frameworków,
* baz danych,
* systemów kolejkowych,
* WebSocketów,

jeżeli nie ma realnej potrzeby.

Preferowana jest prosta architektura klient-serwer, którą można później rozszerzać.

---

# 64. Zasada odpowiedzialności warstw

Przed implementacją nowej funkcjonalności należy określić, do której warstwy należy.

### UI

→ frontend

### Rendering

→ rendering / PixiJS / Three.js

### Interakcja

→ frontend

### Model grafu

→ backend / domain

### Optymalizacja

→ backend / optimization

### Obliczenia

→ backend

### Komunikacja

→ frontend/api lub backend/api

Nie należy umieszczać logiki w niewłaściwej warstwie tylko dlatego, że jest tam łatwiejsza do implementacji.

---

# 65. Zasada separacji modelu i widoku

Szczególnie ważne:

```text
DOMAIN MODEL
```

nie może zależeć od:

```text
RENDERER
```

Czyli:

```text
Graph
```

nie może zawierać:

```text
PIXI.Graphics
PIXI.Container
THREE.Mesh
React Component
```

Analogicznie renderer nie powinien zawierać właściwej logiki optymalizacyjnej.

---

# 66. Zasada źródła prawdy

Backend jest źródłem prawdy dla:

* grafu,
* krawędzi,
* ścian,
* podgrafu,
* problemu optymalizacyjnego,
* rozwiązania,
* statystyk.

Frontend jest źródłem prawdy dla tymczasowego stanu interakcji i UI, np.:

* aktywnego narzędzia,
* pozycji kamery,
* zoomu,
* aktualnego hovera,
* otwartego panelu.

---

# 67. Zasada kompatybilności z 3D

Podczas projektowania nowych elementów należy zadawać sobie pytanie:

> Czy ta decyzja blokuje późniejsze przejście z 2D do 3D?

Nie należy jednak tworzyć niepotrzebnie skomplikowanej architektury tylko po to, aby „być gotowym na 3D”.

Należy przede wszystkim unikać błędnych założeń, takich jak:

```text
Vertex zawsze ma tylko X i Y
```

jeżeli nie jest to konieczne.

Model domenowy powinien umożliwiać późniejsze rozszerzenie do:

```text
X
Y
Z
```

---

# 68. Zasada rozwoju algorytmów

Algorytmy powinny być możliwie niezależne od sposobu wizualizacji.

Powinny móc działać na:

```text
Graph
```

bez uruchamiania:

```text
React
PixiJS
Three.js
Browser
```

Dzięki temu algorytmy będą łatwe do:

* testowania,
* benchmarkowania,
* porównywania,
* uruchamiania na dużych instancjach,
* wykorzystywania w eksperymentach do pracy dyplomowej.

---

# 69. Obecny etap projektu

Aktualnie projekt składa się z:

```text
frontend/
backend/
```

Frontend został utworzony przy użyciu:

```text
Vite
React
TypeScript
```

Backend jest napisany w:

```text
Go
```

Pierwszym celem jest poprawne działanie komunikacji:

```text
React
  │
  │ GET /api/health
  ▼
Go
  │
  ▼
{"status":"ok"}
```

Dopiero po poprawnym działaniu tej komunikacji należy rozwijać właściwe funkcjonalności aplikacji.

---

# 70. Kolejność rozwoju projektu

## Faza 1 — środowisko

* frontend,
* backend,
* komunikacja,
* `/api/health`.

## Faza 2 — podstawowy UI

* top bar,
* side panel,
* canvas,
* status bar.

## Faza 3 — plansza 2D

* PixiJS,
* generowanie kraty,
* pola/wierzchołki,
* sąsiedztwo,
* ściany,
* brakujące połączenia.

## Faza 4 — nawigacja

* pan,
* zoom,
* reset,
* fit-to-screen.

## Faza 5 — interakcja

* wybór pól,
* zaznaczanie,
* edycja połączeń,
* dodawanie ścian,
* tryb Pan,
* tryb Select,
* tryb Edit.

## Faza 6 — obszary

* polygon selection,
* SelectionRegion,
* określenie wierzchołków znajdujących się w obszarze,
* Subgraph.

## Faza 7 — backend domenowy

* Graph,
* Grid,
* Vertex,
* Edge,
* Wall,
* Subgraph,
* Route,
* Solution.

## Faza 8 — komunikacja z backendem

* tworzenie problemu,
* przesyłanie grafu,
* przesyłanie podgrafu,
* przesyłanie ograniczeń,
* pobieranie rozwiązania.

## Faza 9 — pierwszy algorytm

* znajdowanie trasy,
* pokrywanie wierzchołków,
* obliczanie długości,
* uwzględnianie ścian.

## Faza 10 — wiele tras

* ograniczenie liczby tras,
* minimalizacja sumy długości,
* sprawdzanie pokrycia,
* statystyki.

## Faza 11 — kolejne algorytmy

* implementacja kolejnych metod,
* porównywanie,
* benchmarki,
* eksperymenty.

## Faza 12 — zaawansowana wizualizacja

* warstwy,
* animacje,
* proces optymalizacji,
* dodatkowe statystyki.

## Faza 13 — 3D

* Grid3D,
* współrzędne X/Y/Z,
* Three.js,
* kamera 3D,
* pola przestrzenne,
* ściany 3D,
* trasy 3D.

---

# 71. Najważniejsze zasady dla AI

Podczas implementowania każdej nowej funkcjonalności należy brać pod uwagę cały kierunek projektu:

```text
2D → 3D
małe grafy → duże grafy
jeden algorytm → wiele algorytmów
prosty wynik → analiza eksperymentalna
```

Jednocześnie:

**Nie należy implementować przyszłych funkcjonalności przed czasem.**

Należy tworzyć proste abstrakcje, które nie blokują przyszłego rozwoju, ale unikać przedwczesnej komplikacji.

Jeżeli istnieje kilka możliwych rozwiązań, preferowane jest rozwiązanie:

1. proste,
2. czytelne,
3. testowalne,
4. wydajne,
5. modularne,
6. łatwe do rozszerzenia,
7. zgodne z istniejącą architekturą.

---

# 72. Najważniejsza zasada całego projektu

Aplikacja nie jest klasycznym edytorem grafów z punktami i liniami.

Jest **interaktywnym środowiskiem pracy z przestrzenną kratą**, w którym:

```text
KWADRAT / POLE
      ↓
WIERZCHOŁEK GRAFU
```

a:

```text
BRAK PRZESZKODY
      ↓
MOŻLIWOŚĆ PRZEJŚCIA
      ↓
KRAWĘDŹ
```

natomiast:

```text
PUSTA PRZESTRZEŃ / PRZERWA
           ↓
     BRAK POŁĄCZENIA
```

oraz:

```text
GRUBA ŚCIANA / PRZESZKODA
           ↓
     BRAK POŁĄCZENIA
```

Trasy są ciągami przejść pomiędzy dostępnymi polami.

Podgraf jest fragmentem tej przestrzeni.

Obszar zaznaczony przez użytkownika jest geometryczną definicją tego fragmentu.

Algorytm optymalizacyjny działa na rzeczywistym modelu grafu, a frontend wizualizuje jego stan i znalezione rozwiązania.

Cała architektura powinna być projektowana tak, aby w przyszłości ten sam model problemu można było wizualizować zarówno jako **kratę 2D w PixiJS**, jak i **kratę 3D w Three.js**.
