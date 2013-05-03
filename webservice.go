package webservice

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type SoapIn interface {
	GetAction() string
}

type SoapEnvelope struct {
	XMLName  xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	AttrXsi  string   `xml:"xmlns xsi,attr"`
	AttrXsd  string   `xml:"xmlns xsd,attr"`
	AttrSoap string   `xml:"xmlns soap,attr"`
	Body     SoapBody
}

func NewSoapEnvelope() *SoapEnvelope {
	se := &SoapEnvelope{}
	se.AttrXsi = "http://www.w3.org/2001/XMLSchema-instance"
	se.AttrXsd = "http://www.w3.org/2001/XMLSchema"
	se.AttrSoap = "http://schemas.xmlsoap.org/soap/envelope/"

	return se
}

type SoapBody struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
	Content string   `xml:",innerxml"`
}

type SoapFault struct {
	XMLName xml.Name `xml:"Fault"`
	FaultCode string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
	Detail string `xml:"detail"`
}

func CallService(si SoapIn, url string) (sr *SoapEnvelope, err error) {
	// cria o soap envelope
	se := NewSoapEnvelope()

	// gerar o conteúdo do corpo em xml
	bsi, err := xml.Marshal(&si)
	if err != nil {
		return nil, err
	}

	// associa o corpo da requisição
	se.Body.Content = string(bsi)

	// gerar o xml da requisição
	bse, err := xml.Marshal(&se)
	if err != nil {
		return nil, err
	}

	// cria um reader para o corpo da requisição
	br := strings.NewReader(string(bse))
	
	// cria a requisição
	req, err := http.NewRequest("POST", url, br)
	if err != nil {
		return nil, err
	}
	

	// adiciona os cabeçalhos http necessários do soap
	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.Header.Add("SOAPAction", si.GetAction())

	// executa a requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	
	// le o conteudo do retorno
	bsr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// gerar a estrutura de retorno
	err = xml.Unmarshal(bsr, &sr)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 500 {
		var sf SoapFault
		err = xml.Unmarshal([]byte(sr.Body.Content), &sf)
		if err != nil {
			return nil, errors.New(resp.Status)
		}
		return nil, errors.New(sf.FaultString)
	}

	return sr, nil
}
