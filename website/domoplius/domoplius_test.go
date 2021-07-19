package domoplius

import "testing"

type DomopliusData struct {
	Provided string
	Expected string
}

var DomopliustestData = []DomopliusData{
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var z,rr9,wwa,nnj,lly,f,iit,qq5,yym,dd7,xxg,h,v,oo4;
dd7='9';z='5';lly='0';wwa=' ';rr9='7';qq5='9';oo4='6';v='+';xxg='3';yym='7';nnj=' ';f='0';iit='6';h='0'; document.write(v+xxg+rr9+lly+wwa+oo4+dd7+iit+nnj+z+f+h+yym); document.write(qq5);document.write('</span>');</script>            `,
		Expected: "+37069650079",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var h,jji,tt6,a,v,l,b,ssw,m,yyv,g;
a=' ';jji='8';v='7';l='0';g='5';b='7';tt6='6';yyv='3';h='8';m=' ';ssw='0'; document.write(h+a+tt6+v+g+m+ssw+l+yyv+jji+b);document.write('</span>');</script>            `,
		Expected: "+37067500387",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var m,ees,aar,llv,r,xxp,z,nni,d,y,t;
xxp='0';r='6';nni='1';y=' ';m='2';d='5';t=' ';ees='4';llv='6';aar='3';z='8'; document.write(z+t+llv); document.write(m+nni+y+r+xxp+d+ees+aar);document.write('</span>');</script>            `,
		Expected: "+37062160543",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var ggi,ll3,mmu,k,ccn,ii2,bbg,nnv,w,uux,eee,ppl,y,oot;
w='3';nnv='2';ggi='5';ll3='0';ppl='0';bbg='2';uux=' ';y=' ';eee='7';ccn='8';k='0';mmu='6';ii2='+';oot='2'; document.write(ii2+w+eee+ppl+y+mmu+nnv+ll3+uux+ggi+oot+ccn+k); document.write(bbg);document.write('</span>');</script>            `,
		Expected: "+37062052802",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var e,aa1,vvz,m,ooa,ssm,r,nnq,q,b,g,cce,ff4,x;
nnq='3';b='8';x=' ';ssm='7';cce=' ';m='0';q='5';ff4='4';ooa='2';r='8';aa1='6';vvz='+';g='0';e='8'; document.write(vvz); document.write(nnq+ssm+m); document.write(x+aa1+g+r+cce+ooa+ff4+b+q+e);document.write('</span>');</script>            `,
		Expected: "+37060824858",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var vv3,t,b,ll4,aaq,ffm,j,k,yyb,c,e,r,mm9,w;
c=' ';r='0';mm9='0';k='2';t='0';ll4='+';w='2';vv3='5';ffm='3';yyb=' ';aaq='8';e='7';b='2';j='6'; document.write(ll4+ffm+e+t+yyb+j+w+mm9+c+vv3+k+aaq+r+b);document.write('</span>');</script>            `,
		Expected: "+37062052802",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var nni,vvi,hhj,q,k,ccb,rr2,s,w,ppv,b,zzs,l,f;
rr2=' ';nni='4';l='+';f='3';vvi='0';k='5';q='7';ccb='5';ppv='6';w='6';zzs='7';hhj='1';b=' ';s='5'; document.write(l); document.write(f+zzs+vvi+rr2); document.write(w); document.write(q+s+b+ppv+ccb+hhj+nni); document.write(k);document.write('</span>');</script>            `,
		Expected: "+37067565145",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var tth,z,eeo,hhk,ddk,p,ffk,r,kkg,l,ccy,gg5,qqr,u;
z='6';p='5';l='7';ffk='3';ccy='3';hhk='+';eeo='5';r='9';u='7';tth=' ';gg5='8';ddk='3';qqr='0';kkg=' '; document.write(hhk+ddk+l+qqr); document.write(tth+z+eeo+u+kkg+gg5+r+ffk+ccy+p);document.write('</span>');</script>            `,
		Expected: "+37065789335",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var bbg,f,t,ee9,iiy,qqy,a,h,o,rrg,s,kkk,vvb,yym;
t=' ';qqy=' ';o='8';a='0';bbg='3';s='3';rrg='7';f='6';iiy='0';kkk='5';ee9='4';vvb='+';yym='7';h='3'; document.write(vvb+h+rrg); document.write(a+t+f+iiy+yym+qqy+bbg+o+kkk+s+ee9);document.write('</span>');</script>            `,
		Expected: "+37060738534",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var q,lle,ii2,vvd,z,u,ppc,wwo,s,tto,kkv,aa3,jjp,ddc;
s='1';aa3='6';jjp='3';ppc='7';vvd='6';wwo='6';lle='0';z=' ';ddc='6';u=' ';ii2='1';q='+';tto='1';kkv='6'; document.write(q+jjp+ppc+lle+z+kkv+vvd+ddc+u+ii2+tto+aa3+wwo+s);document.write('</span>');</script>            `,
		Expected: "+37066611661",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var r,oox,mmn,dd9,g,bbd,kk6,nnw,ssg,t,p,yyf,w,uux;
uux='6';bbd='1';kk6=' ';dd9='1';t='7';nnw='6';mmn='0';g='6';yyf='+';oox='3';ssg=' ';p='6';w='1';r='6'; document.write(yyf+oox); document.write(t+mmn+kk6); document.write(nnw+g+p); document.write(ssg+w+bbd); document.write(uux+r); document.write(dd9);document.write('</span>');</script>            `,
		Expected: "+37066611661",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var c,ll2,k,n,ii6,b,ssu,z,a,h,pp3,ooh,q,rrz;
ooh='6';k='1';ll2='6';pp3='6';rrz='7';q='1';c=' ';b='+';ssu='6';ii6='3';n='6';z='1';a=' ';h='0'; document.write(b); document.write(ii6+rrz); document.write(h+c+pp3+n+ll2+a+k); document.write(q+ssu+ooh+z);document.write('</span>');</script>            `,
		Expected: "+37066611661",
	},
	{
		Provided: `
		<script type="text/javascript">document.write('<span id="cps4">');var kkd,q,www,eez,d,x,cck,t,m,aag,hh2,vvr,p,b;
t=' ';d=' ';kkd='6';vvr='6';www='6';hh2='6';b='1';eez='1';x='0';aag='6';m='3';p='1';q='7';cck='+'; document.write(cck+m+q); document.write(x+t+vvr+kkd+www); document.write(d+b+eez+aag+hh2+p);document.write('</span>');</script>            `,
		Expected: "+37066611661",
	},
}

func TestDomopliusDecodeNumber(t *testing.T) {
	for _, v := range DomopliustestData {
		if res := domopliusDecodeNumber(v.Provided); res != v.Expected {
			t.Errorf("Result is incorrect, got: '%s', want: '%s'.", res, v.Expected)
		}
	}
}
