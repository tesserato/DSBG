# Ideal String Equation

$$\frac{\partial^{2}}{\partial t^{2}} y{\left(x,t \right)} = c^{2} \frac{\partial^{2}}{\partial x^{2}} y{\left(x,t \right)}$$

Where:

- $y{\left(x,t \right)} \left(\text{m}\right)$: vertical displacement
- $c \left(\frac{\text{m}}{\text{s}}\right)=\frac{\sqrt{\text{T}_0}}{\sqrt{\mu}}=2 \text{{L}} f_{0}$: speed of the wave in the string
- $\text{T}_0 \left(\text{N}\right)$: Tension on the string
- $\mu \left(\frac{\text{kg}}{\text{m}}\right)=\rho \text{A}$: linear density of the string
- $\rho \left(\frac{\text{kg}}{\text{m}^{3}}\right)$: mass density of the string
- $\text{A} \left(\text{m}^{2}\right)$: cross-sectional area of the string
- $\text{E} \left(\frac{\text{N}}{\text{m}^{2}}\right)$: Young's modulus of the string
- $\text{I} \left(\text{kg} \,.\, \text{m}^{2}\right)$: bar moment of inertia
- $\kappa \left(\sqrt{\frac{\text{N}}{\text{m}^{3}}}\right)=\frac{\sqrt{\text{E}} \sqrt{\text{I}}}{\sqrt{\mu} \text{{L}}^{2}}$: bar stiffness
- **Units**:
  - $\text{N} \left(\text{ newton }\right)$ 
  - $\text{kg} \left(\text{ kilogram }\right)$ 
  - $\text{m} \left(\text{ meter }\right)$ 
  - $\text{s} \left(\text{ second }\right)$ 


## Difference Equation

$$- \frac{2 y{\left(x,t \right)}}{{\Delta}t^{2}} + \frac{y{\left(x,t - {\Delta}t \right)}}{{\Delta}t^{2}} + \frac{y{\left(x,t + {\Delta}t \right)}}{{\Delta}t^{2}} = c^{2} \left(- \frac{2 y{\left(x,t \right)}}{{\Delta}x^{2}} + \frac{y{\left(x - {\Delta}x,t \right)}}{{\Delta}x^{2}} + \frac{y{\left(x + {\Delta}x,t \right)}}{{\Delta}x^{2}}\right)$$

Discretizing with:

- $t=\text{n} {\Delta}t,$ where $\text{n}=0,1,2,\cdots,\text{N}-1$
- $x=\text{i} {\Delta}x,$ where $\text{i}=0,1,2,\cdots,\text{M}-1$

We get:

$$\frac{{y}_{\text{i},\text{n} + 1}}{{\Delta}t^{2}} + \frac{{y}_{\text{i},\text{n} - 1}}{{\Delta}t^{2}} - \frac{2 {y}_{\text{i},\text{n}}}{{\Delta}t^{2}} = c^{2} \left(\frac{{y}_{\text{i} + 1,\text{n}}}{{\Delta}x^{2}} + \frac{{y}_{\text{i} - 1,\text{n}}}{{\Delta}x^{2}} - \frac{2 {y}_{\text{i},\text{n}}}{{\Delta}x^{2}}\right)$$

Solving for ${y}_{\text{i},\text{n} + 1}$:

$${y}_{\text{i},\text{n} + 1} = \frac{c^{2} {\Delta}t^{2} \left({y}_{\text{i} + 1,\text{n}} + {y}_{\text{i} - 1,\text{n}} - 2 {y}_{\text{i},\text{n}}\right)}{{\Delta}x^{2}} - {y}_{\text{i},\text{n} - 1} + 2 {y}_{\text{i},\text{n}}$$

## Independent Variables - we want to control the simulation specifying the following variables:

- $f_{0} \left(\text{Hz}\right)$: natural frequency of the string or bar
- $f_{s} \left(\text{Hz}\right)$: sampling frequency
- $\text{{L}} \left(\text{m}\right)$: length of the string or bar
- $d \left(\text{s}\right)$: duration of the simulation
- **Units**:
  - $\text{Hz} \left(\text{ hertz }\right)$ 
  - $\text{m} \left(\text{ meter }\right)$ 
  - $\text{s} \left(\text{ second }\right)$ 


## Dependent Variables:

- $\text{N} \left(\text{{unitless}}\right) =\left\lfloor{d f_{s}}\right\rfloor$: number of time steps
- $\text{M} \left(\text{{unitless}}\right) =\left\lfloor{\frac{\text{{L}}}{{\Delta}x}}\right\rfloor$: number of space steps
- ${\Delta}t \left(\text{s}\right)=\frac{d}{\left\lfloor{d f_{s}}\right\rfloor}=\frac{d}{\text{N}}$: Delta time
- ${\Delta}x \left(\text{m}\right)$: Delta length
- **Units**:
  - $\text{m} \left(\text{ meter }\right)$ 
  - $\text{s} \left(\text{ second }\right)$ 


#### ${\Delta}x$: Von Neumann stability analysis on the finite difference scheme.
Ansatz: ${y}_{\text{i},\text{n}} = \text{Y} e^{i \left(\text{i} k {\Delta}x + \text{n} w {\Delta}t\right)}$

After inserting the test solution:

$$\text{Y} e^{i w {\Delta}t} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t} = \frac{\text{Y} c^{2} {\Delta}t^{2} e^{i k {\Delta}x} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{2}} - \frac{2 \text{Y} c^{2} {\Delta}t^{2} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{2}} + \frac{\text{Y} c^{2} {\Delta}t^{2} e^{- i k {\Delta}x} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}}{{\Delta}x^{2}} + 2 \text{Y} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t} - \text{Y} e^{- i w {\Delta}t} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}$$

Dividing by the common factor $\text{Y} e^{i \text{i} k {\Delta}x} e^{i \text{n} w {\Delta}t}$, rearranging and simplifying gives:

$$0 = \frac{c^{2} {\Delta}t^{2} e^{i k {\Delta}x}}{{\Delta}x^{2}} - \frac{2 c^{2} {\Delta}t^{2}}{{\Delta}x^{2}} + \frac{c^{2} {\Delta}t^{2} e^{- i k {\Delta}x}}{{\Delta}x^{2}} - e^{i w {\Delta}t} + 2 - e^{- i w {\Delta}t}$$

$$0 = \frac{c^{2} {\Delta}t^{2} \left(\cos{\left(k {\Delta}x \right)} - 1\right)}{{\Delta}x^{2}} - \cos{\left(w {\Delta}t \right)} + 1$$

$$\cos{\left(w {\Delta}t \right)} = \frac{c^{2} {\Delta}t^{2} \left(\cos{\left(k {\Delta}x \right)} - 1\right)}{{\Delta}x^{2}} + 1$$

The two solutions are:

$$w = \left[ \frac{- \operatorname{acos}{\left(\frac{c^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - c^{2} {\Delta}t^{2} + {\Delta}x^{2}}{{\Delta}x^{2}} \right)} + 2 \pi}{{\Delta}t}, \  \frac{\operatorname{acos}{\left(\frac{c^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - c^{2} {\Delta}t^{2} + {\Delta}x^{2}}{{\Delta}x^{2}} \right)}}{{\Delta}t}\right]$$

Since $w \in \mathbb{R}$:

$$-1 \le \frac{c^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - c^{2} {\Delta}t^{2} + {\Delta}x^{2}}{{\Delta}x^{2}} \le 1$$

$$- 2 {\Delta}x^{2} \le c^{2} {\Delta}t^{2} \cos{\left(k {\Delta}x \right)} - c^{2} {\Delta}t^{2} \le 0$$

$$- \frac{2 {\Delta}x^{2}}{c^{2} {\Delta}t^{2}} \le\cos{\left(k {\Delta}x \right)} - 1 \le 0$$

$$- \frac{2 {\Delta}x^{2}}{c^{2} {\Delta}t^{2}} \le -2 $$

$$- \frac{{\Delta}x^{2}}{c^{2} {\Delta}t^{2}} \le -1$$

$$\frac{{\Delta}x^{2}}{c^{2} {\Delta}t^{2}} \ge 1$$

$$\frac{{\Delta}x}{c {\Delta}t} \ge 1$$

$${\Delta}x \ge c {\Delta}t$$

$${\Delta}x \ge 2 \text{{L}} f_{0} {\Delta}t$$

