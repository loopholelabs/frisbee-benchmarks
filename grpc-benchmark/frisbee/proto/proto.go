// Code generated by Frisbee v0.1.0, DO NOT EDIT.
// source: benchmark.proto

package proto

import (
	"context"
	"crypto/tls"
	"github.com/loopholelabs/frisbee"
	"github.com/loopholelabs/frisbee/pkg/packet"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"sync"
	"sync/atomic"
)

var (
	NilDecode = errors.New("cannot decode into a nil root struct")
)

type Request struct {
	error  error
	ignore bool

	Message string
}

func NewRequest() *Request {
	return &Request{}
}

func (x *Request) Error(p *packet.Packet, err error) {
	packet.Encoder(p).Error(err)
}

func (x *Request) Encode(p *packet.Packet) {
	if x == nil {
		packet.Encoder(p).Nil()
	} else if x.error != nil {
		packet.Encoder(p).Error(x.error)
	} else {
		packet.Encoder(p).Bool(x.ignore).String(x.Message)
	}
}

func (x *Request) Decode(p *packet.Packet) error {
	if x == nil {
		return NilDecode
	}
	d := packet.GetDecoder(p)
	return x.decode(d)
}

func (x *Request) decode(d *packet.Decoder) error {
	if d.Nil() {
		return nil
	}
	var err error
	x.error, err = d.Error()
	if err != nil {
		x.ignore, err = d.Bool()
		if err != nil {
			return err
		}
		x.Message, err = d.String()
		if err != nil {
			return err
		}
	}
	d.Return()
	return nil
}

type Response struct {
	error  error
	ignore bool

	Message string
}

func NewResponse() *Response {
	return &Response{}
}

func (x *Response) Error(p *packet.Packet, err error) {
	packet.Encoder(p).Error(err)
}

func (x *Response) Encode(p *packet.Packet) {
	if x == nil {
		packet.Encoder(p).Nil()
	} else if x.error != nil {
		packet.Encoder(p).Error(x.error)
	} else {
		packet.Encoder(p).Bool(x.ignore).String(x.Message)
	}
}

func (x *Response) Decode(p *packet.Packet) error {
	if x == nil {
		return NilDecode
	}
	d := packet.GetDecoder(p)
	return x.decode(d)
}

func (x *Response) decode(d *packet.Decoder) error {
	if d.Nil() {
		return nil
	}
	var err error
	x.error, err = d.Error()
	if err != nil {
		x.ignore, err = d.Bool()
		if err != nil {
			return err
		}
		x.Message, err = d.String()
		if err != nil {
			return err
		}
	}
	d.Return()
	return nil
}

type BenchmarkService interface {
	Benchmark(context.Context, *Request) (*Response, error)
	BenchmarkSlow(context.Context, *Request) (*Response, error)
}

type Server struct {
	*frisbee.Server
}

func NewServer(benchmarkService BenchmarkService, listenAddr string, tlsConfig *tls.Config, logger *zerolog.Logger) (*Server, error) {
	table := make(frisbee.HandlerTable)
	table[10] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewRequest()
		err := req.Decode(incoming)
		if err == nil {
			if req.ignore {
				benchmarkService.Benchmark(ctx, req)
			} else {
				var res *Response
				outgoing = incoming
				outgoing.Content.Reset()
				res, err = benchmarkService.Benchmark(ctx, req)
				if err != nil {
					res.Error(outgoing, err)
				} else {
					res.Encode(outgoing)
				}
				outgoing.Metadata.ContentLength = uint32(len(outgoing.Content.B))
			}
		}
		return
	}
	table[11] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		req := NewRequest()
		err := req.Decode(incoming)
		if err == nil {
			if req.ignore {
				benchmarkService.BenchmarkSlow(ctx, req)
			} else {
				var res *Response
				outgoing = incoming
				outgoing.Content.Reset()
				res, err = benchmarkService.BenchmarkSlow(ctx, req)
				if err != nil {
					res.Error(outgoing, err)
				} else {
					res.Encode(outgoing)
				}
				outgoing.Metadata.ContentLength = uint32(len(outgoing.Content.B))
			}
		}
		return
	}
	var s *frisbee.Server
	var err error
	if tlsConfig != nil {
		s, err = frisbee.NewServer(listenAddr, table, frisbee.WithTLS(tlsConfig), frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	} else {
		s, err = frisbee.NewServer(listenAddr, table, frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	}
	return &Server{
		Server: s,
	}, nil
}

type Client struct {
	*frisbee.Client
	nextBenchmark           atomic.Value
	inflightBenchmarkMu     sync.RWMutex
	inflightBenchmark       map[uint16]chan *Response
	nextBenchmarkSlow       atomic.Value
	inflightBenchmarkMuSlow sync.RWMutex
	inflightBenchmarkSlow   map[uint16]chan *Response
}

func NewClient(addr string, tlsConfig *tls.Config, logger *zerolog.Logger) (*Client, error) {
	c := new(Client)
	table := make(frisbee.HandlerTable)
	table[10] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.inflightBenchmarkMu.RLock()
		if ch, ok := c.inflightBenchmark[incoming.Metadata.Id]; ok {
			c.inflightBenchmarkMu.RUnlock()
			res := NewResponse()
			res.Decode(incoming)
			ch <- res
		} else {
			c.inflightBenchmarkMu.RUnlock()
		}
		return
	}
	table[11] = func(ctx context.Context, incoming *packet.Packet) (outgoing *packet.Packet, action frisbee.Action) {
		c.inflightBenchmarkMuSlow.RLock()
		if ch, ok := c.inflightBenchmarkSlow[incoming.Metadata.Id]; ok {
			c.inflightBenchmarkMuSlow.RUnlock()
			res := NewResponse()
			res.Decode(incoming)
			ch <- res
		} else {
			c.inflightBenchmarkMuSlow.RUnlock()
		}
		return
	}
	var err error
	if tlsConfig != nil {
		c.Client, err = frisbee.NewClient(addr, table, context.Background(), frisbee.WithTLS(tlsConfig), frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	} else {
		c.Client, err = frisbee.NewClient(addr, table, context.Background(), frisbee.WithLogger(logger))
		if err != nil {
			return nil, err
		}
	}
	c.nextBenchmark.Store(uint16(0))
	c.inflightBenchmark = make(map[uint16]chan *Response)
	c.nextBenchmarkSlow.Store(uint16(0))
	c.inflightBenchmarkSlow = make(map[uint16]chan *Response)
	return c, nil
}

func (c *Client) Benchmark(ctx context.Context, req *Request) (res *Response, err error) {
	ch := make(chan *Response, 1)
	p := packet.Get()
	p.Metadata.Operation = 10
LOOP:
	p.Metadata.Id = c.nextBenchmark.Load().(uint16)
	if !c.nextBenchmark.CompareAndSwap(p.Metadata.Id, p.Metadata.Id+1) {
		goto LOOP
	}
	req.Encode(p)
	p.Metadata.ContentLength = uint32(len(p.Content.B))
	c.inflightBenchmarkMu.Lock()
	c.inflightBenchmark[p.Metadata.Id] = ch
	c.inflightBenchmarkMu.Unlock()
	err = c.Client.WritePacket(p)
	if err != nil {
		packet.Put(p)
		return
	}
	select {
	case res = <-ch:
		err = res.error
	case <-ctx.Done():
		err = ctx.Err()
	}
	c.inflightBenchmarkMu.Lock()
	delete(c.inflightBenchmark, p.Metadata.Id)
	c.inflightBenchmarkMu.Unlock()
	packet.Put(p)
	return
}

func (c *Client) BenchmarkSlow(ctx context.Context, req *Request) (res *Response, err error) {
	ch := make(chan *Response, 1)
	p := packet.Get()
	p.Metadata.Operation = 11
LOOP:
	p.Metadata.Id = c.nextBenchmarkSlow.Load().(uint16)
	if !c.nextBenchmarkSlow.CompareAndSwap(p.Metadata.Id, p.Metadata.Id+1) {
		goto LOOP
	}
	req.Encode(p)
	p.Metadata.ContentLength = uint32(len(p.Content.B))
	c.inflightBenchmarkMuSlow.Lock()
	c.inflightBenchmarkSlow[p.Metadata.Id] = ch
	c.inflightBenchmarkMuSlow.Unlock()
	err = c.Client.WritePacket(p)
	if err != nil {
		packet.Put(p)
		return
	}
	select {
	case res = <-ch:
		err = res.error
	case <-ctx.Done():
		err = ctx.Err()
	}
	c.inflightBenchmarkMuSlow.Lock()
	delete(c.inflightBenchmarkSlow, p.Metadata.Id)
	c.inflightBenchmarkMuSlow.Unlock()
	packet.Put(p)
	return
}

func (c *Client) BenchmarkIgnore(ctx context.Context, req *Request) (err error) {
	p := packet.Get()
	p.Metadata.Operation = 10
LOOP:
	p.Metadata.Id = c.nextBenchmark.Load().(uint16)
	if !c.nextBenchmark.CompareAndSwap(p.Metadata.Id, p.Metadata.Id+1) {
		goto LOOP
	}
	req.ignore = true
	req.Encode(p)
	p.Metadata.ContentLength = uint32(len(p.Content.B))
	err = c.Client.WritePacket(p)
	packet.Put(p)
	return
}
