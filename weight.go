package main

type ClusterWeigth struct {
	weight int
	svc    ServerType
}

type ClusterSelect struct {
	List      []ClusterWeigth
	Idx       int
	Gcd       int
	MaxWeight int
	CurWeight int
}

func NewClusterWeight(list []ClusterWeigth) *ClusterSelect {

	s := &ClusterSelect{List: make([]ClusterWeigth, len(list))}
	for idx, v := range list {
		s.List[idx] = v
		if s.MaxWeight < v.weight {
			s.MaxWeight = v.weight
		}
		if idx == 0 {
			s.Gcd = v.weight
		} else {
			s.Gcd = gcd(s.Gcd, v.weight)
		}
	}
	s.CurWeight = s.MaxWeight

	return s
}

func (s *ClusterSelect) Reset() {
	s.CurWeight = s.MaxWeight
	s.Idx = 0
}

func (s *ClusterSelect) Select() ServerType {

	for {
		s.Idx = (s.Idx + 1) % len(s.List)
		if s.Idx == 0 {
			s.CurWeight = s.CurWeight - s.Gcd
			if s.CurWeight <= 0 {
				s.CurWeight = s.MaxWeight
			}
		}
		if s.List[s.Idx].weight >= s.CurWeight {
			return s.List[s.Idx].svc
		}
	}
}

/* 迭代法（递推法）：欧几里得算法，计算最大公约数 */
func gcd(m, n int) int {
	for {
		if m == 0 {
			return n
		}
		c := n % m
		n = m
		m = c
	}
}
