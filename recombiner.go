package fastjet

import (
	"fmt"
	"math"

	"github.com/go-hep/fmom"
)

type Recombiner interface {
	Description() string
	Recombine(j1, j2 *Jet) (Jet, error)
	Preprocess(jet *Jet) error
	Scheme() RecombinationScheme
}

type DefaultRecombiner struct {
	scheme RecombinationScheme
}

func NewRecombiner(scheme RecombinationScheme) DefaultRecombiner {
	return DefaultRecombiner{
		scheme: scheme,
	}
}

func (rec DefaultRecombiner) Description() string {
	str := rec.scheme.String()
	return str + " scheme recombination"
}

func (rec DefaultRecombiner) Recombine(j1, j2 *Jet) (Jet, error) {
	var err error
	var jet Jet

	w1 := 0.0
	w2 := 0.0

	switch rec.Scheme() {
	case EScheme:
		jet.PxPyPzE = fmom.NewPxPyPzE(
			j1.Px()+j2.Px(),
			j1.Py()+j2.Py(),
			j1.Pz()+j2.Pz(),
			j1.E()+j2.E(),
		)
		return jet, err

	case PtScheme, EtScheme, BIPtScheme:
		w1 = j1.Pt()
		w2 = j2.Pt()

	case Pt2Scheme, Et2Scheme, BIPt2Scheme:
		w1 = j1.Pt2()
		w2 = j2.Pt2()

	default:
		return jet, fmt.Errorf("fastjet.Recombine: invalid recombination scheme (%v)", rec.Scheme())
	}

	pt := j1.Pt() + j2.Pt()
	if pt != 0.0 {
		y := (w1*j1.Rapidity() + w2*j2.Rapidity()) / (w1 + w2)
		phi1 := j1.Phi()
		phi2 := j2.Phi()
		if phi1-phi2 > math.Pi {
			phi2 += 2 * math.Pi
		}
		if phi1-phi2 < -math.Pi {
			phi2 -= 2 * math.Pi
		}
		phi := (w1*phi1 + w2*phi2) / (w1 + w2)
		jet.PxPyPzE = fmom.NewPxPyPzE(
			pt*math.Cos(phi),
			pt*math.Sin(phi),
			pt*math.Sinh(y),
			pt*math.Cosh(y),
		)
	} else {
		jet.PxPyPzE = fmom.NewPxPyPzE(0, 0, 0, 0)
	}

	return jet, err
}

func (rec DefaultRecombiner) Preprocess(jet *Jet) error {
	var err error

	switch rec.Scheme() {
	case EScheme, BIPtScheme, BIPt2Scheme:
		return nil

	case PtScheme, Pt2Scheme:
		// these schemes (as in the ktjet impl.) need massless
		// initial 4-vectors, with essentially E=|p|
		pz := jet.Pz()
		e := math.Sqrt(jet.Pt2() + pz*pz)
		jet.PxPyPzE = fmom.NewPxPyPzE(jet.Px(), jet.Py(), jet.Pz(), e)
		return nil

	case EtScheme, Et2Scheme:
		// these schemes (as in the ktjet impl.) need massless
		// initial 4-vectors, with essentially E=|p|
		pz := jet.Pz()
		rescale := jet.E() / math.Sqrt(jet.Pt2()+pz*pz)
		jet.PxPyPzE = fmom.NewPxPyPzE(
			rescale*jet.Px(),
			rescale*jet.Py(),
			rescale*jet.Pz(),
			jet.E(),
		)
		return nil

	default:
		return fmt.Errorf("fastjet.Preprocess: invalid recombination scheme (%v)", rec.Scheme())
	}

	return err
}

func (rec DefaultRecombiner) Scheme() RecombinationScheme {
	return rec.scheme
}
