/*eslint-env node, es6, jasmine */
/*eslint indent: ['error', 2], quotes: [0, 'single'] */
const sinon = require('sinon');
const chai = require('chai');
const expect = chai.expect;


describe('Treetop', function() {
  'strict mode';
  var requests;
  var treetop;

  beforeEach(function() {
    this.xhr = sinon.useFakeXMLHttpRequest();
    global.XMLHttpRequest = this.xhr;
    requests = [];
    treetop = window.treetop;
    this.xhr.onCreate = req => requests.push(req);
    treetop.push();
    window.requestAnimationFrame.lastCall.args[0]();
  });

  afterEach(function() {
    this.xhr.restore();
    window.requestAnimationFrame.reset();
    window.cancelAnimationFrame.reset();
  });

  describe('issue basic GET request', function() {
    var req = null;
    beforeEach(function() {
      treetop.request("GET", "/test");
      req = requests[0];});

    it('should have issued a request', () => expect(req).to.exist);

    it('should have issued a request with the method and url', function() {
      expect(req.url).to.contain("/test");
      expect(req.method).to.equal("GET");
    });

    it('should have added the treetop header', () => expect(req.requestHeaders["accept"]).to.contain(treetop.PARTIAL_CONTENT_TYPE));

    it('should have no body', () => expect(req.requestBody).to.be.null);
  });

  describe('issue basic POST request', function() {
    var req = null;
    beforeEach(function() {
      treetop.request("POST", "/test", "a=123&b=987", "application/x-www-form-urlencoded");
      req = requests[0];
    });

    it('should have issued a request with right info', function() {
      expect(req).to.exist;
      expect(req.url).to.contain("/test");
      expect(req.method).to.equal("POST");
      expect(req.requestHeaders["accept"]).to.contain(treetop.PARTIAL_CONTENT_TYPE);
      expect(req.url).to.equal("/test");
    });

    it('should have added the content type header', () =>
      expect(req.requestHeaders["content-type"])
        .to.contain("application/x-www-form-urlencoded")
    );

    it('should have a body', () => expect(req.requestBody).to.equal("a=123&b=987"));
  });

  describe('rejected request', () =>
    it('should have a white list of methods', () =>
      expect(() => treetop.request("NOMETHOD"))
        .to.throw("Treetop: Unknown request method 'NOMETHOD'")
    )
  );

  describe('replace indexed elements', function() {
    var el = null;
    beforeEach(function() {
      el = document.createElement("p");
      el.setAttribute("id", "test");
      el.textContent = "before!";
      document.body.appendChild(el);
    });

    afterEach(() => document.body.removeChild(document.getElementById("test")));

    it('should have appended the child', () => expect(el.parentNode.tagName).to.equal("BODY"));

    it('should replace <p>before!</p> with <em>after!</em>', function() {
      treetop.request("GET", "/test");
      requests[0].respond(
        200,
        { 'content-type': treetop.PARTIAL_CONTENT_TYPE },
        '<em id="test">after!</em>'
      );
      expect(document.body.textContent).to.equal("after!");
    });

    it('should do nothing with an unmatched response', function() {
      treetop.request("GET", "/test");
      requests[0].respond(
        200,
        { 'content-type': treetop.PARTIAL_CONTENT_TYPE },
        '<em id="test_other">after!</em>'
      );
      expect(document.body.textContent).to.equal("before!");
    });
  });

  describe('replace singleton elements', () =>

    it('should replace title tag', function() {
      treetop.request("GET", "/test");
      requests[0].respond(
        200,
        { 'content-type': treetop.PARTIAL_CONTENT_TYPE },
        '<title>New Title!</title>'
      );
      expect(document.title).to.equal("New Title!");
    })
  );

  describe('mounting and unmounting elements', function() {
    beforeEach(function() {
      this.el = document.createElement("DIV");
      this.el.textContent = "Before!";
      this.el.setAttribute("id", "test");
      document.body.appendChild(this.el);
      return treetop.mount(document.body);
    });

    afterEach(() => document.body.removeChild(document.getElementById("test")));

    it('should have mounted the body element', () => expect(document.body._treetopComponents).to.exist);

    it('should have mounted the child element', function() {
      expect(this.el._treetopComponents).to.eql([]);
    });

    describe('when elements are replaced', function() {
      beforeEach(function() {
        treetop.request("GET", "/test");
        requests[0].respond(
          200,
          { 'content-type': treetop.PARTIAL_CONTENT_TYPE },
          '<em id="test">after!</em>'
        );
        return this.nue = document.getElementById('test');
      });

      it('should remove the element from the DOM', function() {
        expect(this.el.parentNode).to.be.null;
      });

      it('should unmount the existing element', function() {
        expect(this.el._treetopComponents).to.be.null;
      });

      it('should have inserted the new #test element', function() {
        expect(this.nue.tagName).to.equal("EM");
      });

      it('should mount the new element', function() {
        expect(this.nue._treetopComponents).to.eql([]);
      });
    });
  });

  describe('binding components', function() {
    beforeEach(function() {
      this.el = document.createElement("test-node");
      this.el.setAttribute("id", "test");
      document.body.appendChild(this.el);
      this.el2 = document.createElement("div");
      this.el2.setAttribute("id", "test2");
      this.el2.setAttribute("test-node", 123);
      document.body.appendChild(this.el2);
      // component definition:
      this.component = {
        tagName: "test-node",
        attrName: "test-node",
        mount: sinon.spy(),
        unmount: sinon.spy()
      };
      treetop.push(this.component);
      window.requestAnimationFrame.lastCall.args[0]();
    });

    afterEach(function() {
      document.body.removeChild(document.getElementById("test"));
      document.body.removeChild(document.getElementById("test2"));
    });

    it('should have called the mount on the element', function() {
      expect(this.component.mount.calledWith(this.el)).to.be.true;
    });

    it('should have called the mount on the attribute', function() {
      expect(this.component.mount.calledWith(this.el2)).to.be.true;
    });

    describe('when unmounted', function() {
      beforeEach(function() {
        treetop.request("GET", "/test");
        requests[0].respond(
          200,
          { 'content-type': treetop.PARTIAL_CONTENT_TYPE },
          '<div id="test">after!</div><div id="test2">after2!</div>'
        );
      });

      it('should have called the unmount on the element', function() {
        expect(this.component.unmount.calledWith(this.el)).to.be.true;
      });

      it('should have called the unmount on the attribute', function() {
        expect(this.component.unmount.calledWith(this.el2)).to.be.true;
      });
    });
  });

  describe('binding two components', function() {
    beforeEach(function() {
      this.el = document.createElement("test-node");
      this.el.setAttribute("id", "test");
      document.body.appendChild(this.el);
      this.el2 = document.createElement("div");
      this.el2.setAttribute("id", "test2");
      this.el2.setAttribute("test-node", 123);
      document.body.appendChild(this.el2);
      // component definition:
      this.component = {
        tagName: "test-node",
        attrName: "test-node",
        mount: sinon.spy(),
        unmount: sinon.spy()
      };
      this.component2 = {
        tagName: "test-node",
        attrName: "test-node",
        mount: sinon.spy(),
        unmount: sinon.spy()
      };
      treetop.push(this.component);
      treetop.push(this.component2);
      window.requestAnimationFrame.lastCall.args[0]();
    });

    afterEach(function() {
      document.body.removeChild(document.getElementById("test"));
      document.body.removeChild(document.getElementById("test2"));
    });

    it('should have called mount on component 1 for the tagName', function() {
      expect(this.component.mount.calledWith(this.el)).to.be.true;
    });

    it('should have called mount on component 1 for the attrName', function() {
      expect(this.component.mount.calledWith(this.el2)).to.be.true;
    });

    it('should have called mount on component 2 for the tagName', function() {
      expect(this.component2.mount.calledWith(this.el)).to.be.true;
    });

    it('should have called mount on component 2 for the attrName', function() {
      expect(this.component2.mount.calledWith(this.el2)).to.be.true;
    });
  });
});

