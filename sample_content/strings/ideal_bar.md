# Euler-Lagrange equation for beams:

$$
 \frac{\partial^{2}}{\partial t^{2}} y{\left(x,t \right)} = - \kappa^{2} \frac{\partial^{4}}{\partial x^{4}} y{\left(x,t \right)} 
$$



# Finite difference:

$$
- \frac{2 y{\left(x,t \right)}}{{\Delta}t^{2}} + \frac{y{\left(x,t - {\Delta}t \right)}}{{\Delta}t^{2}} + \frac{y{\left(x,t + {\Delta}t \right)}}{{\Delta}t^{2}} = - \kappa^{2} \left(\frac{6 y{\left(x,t \right)}}{{\Delta}x^{4}} + \frac{y{\left(x - 2 {\Delta}x,t \right)}}{{\Delta}x^{4}} - \frac{4 y{\left(x - {\Delta}x,t \right)}}{{\Delta}x^{4}} - \frac{4 y{\left(x + {\Delta}x,t \right)}}{{\Delta}x^{4}} + \frac{y{\left(x + 2 {\Delta}x,t \right)}}{{\Delta}x^{4}}\right)
$$

## Discretizing with:
- $t_{n}=\text{n} {\Delta}t,$ where $\text{n}=0,1,2, ..., N$
- $x_{i}=\text{i} {\Delta}x,$ where $\text{i}=0,1,2, ..., M$
$$\frac{{y}_{\text{i},\text{n} + 1}}{{\Delta}t^{2}} + \frac{{y}_{\text{i},\text{n} - 1}}{{\Delta}t^{2}} - \frac{2 {y}_{\text{i},\text{n}}}{{\Delta}t^{2}} = - \kappa^{2} \left(- \frac{4 {y}_{\text{i} + 1,\text{n}}}{{\Delta}x^{4}} + \frac{{y}_{\text{i} + 2,\text{n}}}{{\Delta}x^{4}} - \frac{4 {y}_{\text{i} - 1,\text{n}}}{{\Delta}x^{4}} + \frac{{y}_{\text{i} - 2,\text{n}}}{{\Delta}x^{4}} + \frac{6 {y}_{\text{i},\text{n}}}{{\Delta}x^{4}}\right)$$
# Solving for the next time step we have the finite difference schema:
$${y}_{\text{i},\text{n} + 1} = \frac{\kappa^{2} {\Delta}t^{2} \left(4 {y}_{\text{i} + 1,\text{n}} - {y}_{\text{i} + 2,\text{n}} + 4 {y}_{\text{i} - 1,\text{n}} - {y}_{\text{i} - 2,\text{n}} - 6 {y}_{\text{i},\text{n}}\right)}{{\Delta}x^{4}} - {y}_{\text{i},\text{n} - 1} + 2 {y}_{\text{i},\text{n}}$$
## Independent variables:
- $f_{0} \left(\text{Hz}\right)$: natural frequency of the string or bar
- $f_{s} \left(\text{Hz}\right)$: sampling frequency
- $\text{{L}} \left(\text{m}\right)$: length of the string or bar
- $d \left(\text{s}\right)$: duration of the simulation
- **Units**:
  - $\text{Hz} \left(\text{ hertz }\right)$ 
  - $\text{m} \left(\text{ meter }\right)$ 
  - $\text{s} \left(\text{ second }\right)$ 

## Dependent variables:
- $\text{N} \left(\text{{unitless}}\right) =\left\lfloor{d f_{s}}\right\rfloor$: number of time steps
- $\text{M} \left(\text{{unitless}}\right) =\left\lfloor{\frac{\text{{L}}}{{\Delta}x}}\right\rfloor$: number of space steps
- ${\Delta}t \left(\text{s}\right)=\frac{d}{\left\lfloor{d f_{s}}\right\rfloor}=\frac{d}{\text{N}}$: Delta time
- ${\Delta}x \left(\text{m}\right)$: Delta length
- **Units**:
  - $\text{m} \left(\text{ meter }\right)$ 
  - $\text{s} \left(\text{ second }\right)$ 

### $\kappa$ and ${\Delta}x$ are unknown in terms of the independent variables 
#### $\kappa$: Since, for the simply supported boundary condition:
$f_0 = 22.3733 \sqrt{\frac{\text{E} \text{I}}{\mu \text{{L}}^{4}}}$

And recalling that:

$\kappa = \sqrt{\frac{\text{E} \text{I}}{\mu \text{{L}}^{4}}}$

then:

$f_0 = 22.3733*\kappa$

$\kappa = f_0 / 22.3733$

#### ${\Delta}x$: Von Neumann stability analysis on the finite difference scheme for the Euler-Bernoulli beam equation.
Ansatz: $y_{i,n} = \text{Y} e^{i \left(\text{i} k {\Delta}x + \text{n} w {\Delta}t\right)}$

Replacing the ansatz in the finite difference schema and simplifying:

$$\text{Y} e^{i \left(\text{i} k {\Delta}x + w {\Delta}t \left(\text{n} + 1\right)\right)} = - \frac{\kappa^{2} \text{Y} {\Delta}t^{2} e^{2 i k {\Delta}x} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{4}} + \frac{4 \kappa^{2} \text{Y} {\Delta}t^{2} e^{i k {\Delta}x} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{4}} - \frac{6 \kappa^{2} \text{Y} {\Delta}t^{2} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{4}} + \frac{4 \kappa^{2} \text{Y} {\Delta}t^{2} e^{- i k {\Delta}x} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{4}} - \frac{\kappa^{2} \text{Y} {\Delta}t^{2} e^{- 2 i k {\Delta}x} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{4}} + 2 \text{Y} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t} - \text{Y} e^{- i w {\Delta}t} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}$$

$$e^{i w {\Delta}t} + e^{- i w {\Delta}t} = - \frac{\kappa^{2} {\Delta}t^{2} e^{2 i k {\Delta}x}}{{\Delta}x^{4}} + \frac{4 \kappa^{2} {\Delta}t^{2} e^{i k {\Delta}x}}{{\Delta}x^{4}} - \frac{6 \kappa^{2} {\Delta}t^{2}}{{\Delta}x^{4}} + \frac{4 \kappa^{2} {\Delta}t^{2} e^{- i k {\Delta}x}}{{\Delta}x^{4}} - \frac{\kappa^{2} {\Delta}t^{2} e^{- 2 i k {\Delta}x}}{{\Delta}x^{4}} + 2$$

$$2 \cos{\left(w {\Delta}t \right)} = - \frac{4 \kappa^{2} {\Delta}t^{2} \cos^{2}{\left(k {\Delta}x \right)}}{{\Delta}x^{4}} + \frac{8 \kappa^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)}}{{\Delta}x^{4}} - \frac{4 \kappa^{2} {\Delta}t^{2}}{{\Delta}x^{4}} + 2$$

Solving for $w \in \mathbb{R}$:

$$w = \left[ \frac{- \operatorname{acos}{\left(\frac{- 2 \kappa^{2} {\Delta}t^{2} \cos^{2}{\left(k {\Delta}x \right)} + 4 \kappa^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - 2 \kappa^{2} {\Delta}t^{2} + {\Delta}x^{4}}{{\Delta}x^{4}} \right)} + 2 \pi}{{\Delta}t}, \  \frac{\operatorname{acos}{\left(\frac{- 2 \kappa^{2} {\Delta}t^{2} \cos^{2}{\left(k {\Delta}x \right)} + 4 \kappa^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - 2 \kappa^{2} {\Delta}t^{2} + {\Delta}x^{4}}{{\Delta}x^{4}} \right)}}{{\Delta}t}\right]$$

Thus:

$$ -1 \le \frac{- 2 \kappa^{2} {\Delta}t^{2} \cos^{2}{\left(k {\Delta}x \right)} + 4 \kappa^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - 2 \kappa^{2} {\Delta}t^{2} + {\Delta}x^{4}}{{\Delta}x^{4}} \le 1 $$

$$ - 2 {\Delta}x^{4} \le - 2 \kappa^{2} {\Delta}t^{2} \cos^{2}{\left(k {\Delta}x \right)} + 4 \kappa^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - 2 \kappa^{2} {\Delta}t^{2} \le 0 $$

$$ - \frac{{\Delta}x^{4}}{\kappa^{2} {\Delta}t^{2}} \le - \cos^{2}{\left(k {\Delta}x \right)} + 2 \cos{\left(k {\Delta}x \right)} - 1 \le 0 $$

$$ \frac{{\Delta}x^{4}}{\kappa^{2} {\Delta}t^{2}} \ge \left(\cos{\left(k {\Delta}x \right)} - 1\right)^{2} \ge 0 $$

$$ \frac{{\Delta}x^{2}}{\kappa {\Delta}t} \ge \left|{\cos{\left(k {\Delta}x \right)} - 1}\right| \ge 0 $$

$$ \frac{{\Delta}x^{2}}{\kappa {\Delta}t} \ge 2 $$

$$ {\Delta}x \ge \sqrt{2 \kappa {\Delta}t} $$

